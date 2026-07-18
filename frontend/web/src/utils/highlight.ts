/** 将文本按关键词拆成片段，用于安全高亮（不用 v-html） */
export interface HighlightPart {
  text: string
  hit: boolean
}

function escapeRegExp(s: string) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function isHan(ch: string) {
  return /[\u4e00-\u9fff]/.test(ch)
}

function isLatinWordChar(ch: string) {
  return /[a-zA-Z0-9]/.test(ch)
}

/**
 * 从关键词中提取「确实出现在菜名里」的片段：
 * - 整词命中
 * - 中文连续子串（优先长的，最短到单字）
 * - 英文按空格分词后命中（不做单字母高亮）
 */
export function collectMatchingTerms(keyword: string, text: string): string[] {
  const kw = keyword.trim()
  if (!kw || !text) return []

  const lowerText = text.toLowerCase()
  const seen = new Set<string>()
  const out: string[] = []

  const add = (raw: string) => {
    const t = raw.trim()
    if (!t) return
    const key = t.toLowerCase()
    if (seen.has(key)) return
    // 英文单字符太噪，跳过
    if (t.length === 1 && isLatinWordChar(t)) return
    seen.add(key)
    out.push(t)
  }

  // 整词
  if (lowerText.includes(kw.toLowerCase())) {
    add(kw)
  }

  // 英文分词
  for (const part of kw.split(/\s+/)) {
    if (part.length >= 2 && lowerText.includes(part.toLowerCase())) {
      add(part)
    }
  }

  // 中文：枚举关键词子串，只保留出现在菜名中的
  const chars = Array.from(kw)
  const hasChinese = chars.some(isHan)
  if (hasChinese) {
    for (let len = chars.length; len >= 1; len--) {
      for (let i = 0; i + len <= chars.length; i++) {
        const slice = chars.slice(i, i + len)
        // 纯英文数字片段交给上面的分词逻辑
        if (slice.every((c) => !isHan(c))) continue
        const sub = slice.join('')
        if (text.includes(sub)) {
          add(sub)
        }
      }
    }
  }

  out.sort((a, b) => b.length - a.length || a.localeCompare(b, 'zh'))
  return out
}

export function splitHighlightParts(text: string, terms: string[]): HighlightPart[] {
  if (!text) return []

  // 若传入的是「搜索关键词」本身（单项），按与菜名的交集提取高亮片段
  const list = terms.length === 1
    ? collectMatchingTerms(terms[0], text)
    : collectMatchingTerms(terms.join(' '), text)

  // 兼容：也把原始 terms 中确实出现在文本里的保留
  const merged = new Set(list.map((t) => t.toLowerCase()))
  for (const t of terms) {
    const raw = t.trim()
    if (!raw) continue
    if (text.toLowerCase().includes(raw.toLowerCase()) && !merged.has(raw.toLowerCase())) {
      if (!(raw.length === 1 && isLatinWordChar(raw))) {
        list.push(raw)
        merged.add(raw.toLowerCase())
      }
    }
    // 对每个 term 再做一次交集提取（支持 highlight_terms 多值）
    for (const m of collectMatchingTerms(raw, text)) {
      if (!merged.has(m.toLowerCase())) {
        list.push(m)
        merged.add(m.toLowerCase())
      }
    }
  }

  list.sort((a, b) => b.length - a.length)
  if (!list.length) return [{ text, hit: false }]

  const pattern = list.map(escapeRegExp).join('|')
  if (!pattern) return [{ text, hit: false }]

  const re = new RegExp(`(${pattern})`, 'gi')
  const parts: HighlightPart[] = []
  let last = 0
  let match: RegExpExecArray | null
  while ((match = re.exec(text)) !== null) {
    if (match.index > last) {
      parts.push({ text: text.slice(last, match.index), hit: false })
    }
    parts.push({ text: match[0], hit: true })
    last = match.index + match[0].length
    if (match[0].length === 0) re.lastIndex++
  }
  if (last < text.length) {
    parts.push({ text: text.slice(last), hit: false })
  }
  return parts.length ? parts : [{ text, hit: false }]
}
