/** 将文本按关键词拆成片段，用于安全高亮（不用 v-html） */
export interface HighlightPart {
  text: string
  hit: boolean
}

function escapeRegExp(s: string) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function collectTerms(terms: string[]): string[] {
  const seen = new Set<string>()
  const out: string[] = []
  const add = (raw: string) => {
    const t = raw.trim()
    if (!t) return
    const key = t.toLowerCase()
    if (seen.has(key)) return
    seen.add(key)
    out.push(t)
    // 英文复数简单去 s，便于 egg / eggs 互命中
    if (/^[a-zA-Z]+s$/i.test(t) && t.length > 3) {
      const stem = t.slice(0, -1)
      const stemKey = stem.toLowerCase()
      if (!seen.has(stemKey)) {
        seen.add(stemKey)
        out.push(stem)
      }
    }
  }

  for (const term of terms) {
    add(term)
    for (const part of term.split(/\s+/)) {
      add(part)
    }
  }

  // 长词优先，避免短词先拆坏长匹配
  out.sort((a, b) => b.length - a.length)
  return out
}

export function splitHighlightParts(text: string, terms: string[]): HighlightPart[] {
  if (!text) return []
  const list = collectTerms(terms)
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
    // 避免零宽死循环
    if (match[0].length === 0) re.lastIndex++
  }
  if (last < text.length) {
    parts.push({ text: text.slice(last), hit: false })
  }
  return parts.length ? parts : [{ text, hit: false }]
}
