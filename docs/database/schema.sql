-- =============================================================================
-- 人间烟火 (Delicious) - MySQL 8.0+ DDL
-- 核心设计：my_recipes 与 recipe_versions 版本控制（不可变版本记录）
-- =============================================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- -----------------------------------------------------------------------------
-- 1. 用户表
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
    `id`            BIGINT UNSIGNED     NOT NULL AUTO_INCREMENT COMMENT '主键',
    `username`      VARCHAR(64)         NOT NULL COMMENT '登录用户名',
    `email`         VARCHAR(128)        DEFAULT NULL COMMENT '邮箱',
    `password_hash` VARCHAR(255)        NOT NULL COMMENT '密码哈希 (bcrypt)',
    `nickname`      VARCHAR(64)         NOT NULL DEFAULT '' COMMENT '显示昵称',
    `avatar_url`    VARCHAR(512)        DEFAULT NULL COMMENT '头像 URL',
    `status`        TINYINT             NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 0=禁用',
    `created_at`    DATETIME(3)         NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`    DATETIME(3)         NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at`    DATETIME(3)         DEFAULT NULL COMMENT '软删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_users_username` (`username`),
    UNIQUE KEY `uk_users_email` (`email`),
    KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- -----------------------------------------------------------------------------
-- 2. 百科菜谱表（标准做法与配料比例，供搜索与基准对比）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `encyclopedia_recipes` (
    `id`              BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT COMMENT '主键',
    `name`            VARCHAR(128)      NOT NULL COMMENT '菜名',
    `description`     TEXT              DEFAULT NULL COMMENT '简介',
    `cover_image_url` VARCHAR(512)      DEFAULT NULL COMMENT '封面图',
    `category`        VARCHAR(64)       DEFAULT NULL COMMENT '分类',
    `tags`            JSON              DEFAULT NULL COMMENT '标签数组 ["家常菜","川菜"]',
    -- 标准配料: [{"name":"五花肉","amount":500,"unit":"g","note":"切块"}]
    `ingredients`     JSON              NOT NULL COMMENT '标准配料与重量',
    -- 标准步骤: [{"order":1,"content":"焯水...","duration_minutes":5}]
    `process_steps`   JSON              NOT NULL COMMENT '标准制作步骤',
    `source`          VARCHAR(128)      DEFAULT NULL COMMENT '来源',
    `view_count`      INT UNSIGNED      NOT NULL DEFAULT 0 COMMENT '浏览次数',
    `created_at`      DATETIME(3)       NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`      DATETIME(3)       NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    KEY `idx_encyclopedia_name` (`name`),
    KEY `idx_encyclopedia_category` (`category`),
    FULLTEXT KEY `ft_encyclopedia_name_desc` (`name`, `description`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='百科菜谱表';

-- -----------------------------------------------------------------------------
-- 3. 我的菜谱主表（通用信息 + 指向当前版本）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `my_recipes` (
    `id`                     BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id`                BIGINT UNSIGNED   NOT NULL COMMENT '所属用户',
    `name`                   VARCHAR(128)      NOT NULL COMMENT '菜名',
    `current_version_id`     BIGINT UNSIGNED   DEFAULT NULL COMMENT '当前生效版本 ID（指向 recipe_versions）',
    `user_rating`            TINYINT UNSIGNED  DEFAULT NULL COMMENT '个人评分 1-5 星',
    `cover_image_url`        VARCHAR(512)      DEFAULT NULL COMMENT '列表封面（通常取当前版本首图）',
    `encyclopedia_recipe_id` BIGINT UNSIGNED   DEFAULT NULL COMMENT '关联百科菜谱（基准对比）',
    `created_at`             DATETIME(3)       NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at`             DATETIME(3)       NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at`             DATETIME(3)       DEFAULT NULL COMMENT '软删除',
    PRIMARY KEY (`id`),
    KEY `idx_my_recipes_user_id` (`user_id`),
    KEY `idx_my_recipes_user_rating` (`user_rating`),
    KEY `idx_my_recipes_created_at` (`created_at`),
    KEY `idx_my_recipes_deleted_at` (`deleted_at`),
    KEY `idx_my_recipes_encyclopedia` (`encyclopedia_recipe_id`),
  CONSTRAINT `fk_my_recipes_user`
        FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_my_recipes_encyclopedia`
        FOREIGN KEY (`encyclopedia_recipe_id`) REFERENCES `encyclopedia_recipes` (`id`)
        ON DELETE SET NULL ON UPDATE CASCADE
    -- 注意: current_version_id 外键在 recipe_versions 表创建后追加（避免循环依赖）
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='我的菜谱主表';

-- -----------------------------------------------------------------------------
-- 4. 菜谱版本详情表（核心：每次修改插入新行，不更新旧版本）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `recipe_versions` (
    `id`             BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT COMMENT '主键',
    `recipe_id`      BIGINT UNSIGNED   NOT NULL COMMENT '所属菜谱 my_recipes.id',
    `version_number` INT UNSIGNED      NOT NULL COMMENT '版本号，从 1 递增',
    -- 配料: [{"name":"生抽","amount":2,"unit":"勺","note":""}]
    `ingredients`    JSON              NOT NULL COMMENT '食材与重量',
    -- 步骤（结构化）: [{"order":1,"content":"...","duration_minutes":null}]
    `process_steps`  JSON              NOT NULL COMMENT '制作步骤',
    `process_text`   TEXT              DEFAULT NULL COMMENT '步骤纯文本（可选，兼容简单录入）',
    -- 多图: ["https://cdn/1.jpg","https://cdn/2.jpg"]
    `images`         JSON              DEFAULT NULL COMMENT '版本图片列表',
    `commit_msg`     VARCHAR(255)      NOT NULL DEFAULT '' COMMENT '修改备注（如：减盐、调整火候）',
    `created_at`     DATETIME(3)       NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_recipe_version` (`recipe_id`, `version_number`),
    KEY `idx_recipe_versions_recipe_id` (`recipe_id`),
    KEY `idx_recipe_versions_created_at` (`created_at`),
    CONSTRAINT `fk_recipe_versions_recipe`
        FOREIGN KEY (`recipe_id`) REFERENCES `my_recipes` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜谱版本详情表';

-- 追加 my_recipes.current_version_id 外键（recipe_versions 已存在）
ALTER TABLE `my_recipes`
    ADD CONSTRAINT `fk_my_recipes_current_version`
        FOREIGN KEY (`current_version_id`) REFERENCES `recipe_versions` (`id`)
        ON DELETE SET NULL ON UPDATE CASCADE;

SET FOREIGN_KEY_CHECKS = 1;

-- =============================================================================
-- 版本写入流程（应用层事务）:
--   1. INSERT my_recipes (current_version_id = NULL)  -- 新建时
--   2. INSERT recipe_versions (version_number = MAX+1 或 1)
--   3. UPDATE my_recipes SET current_version_id = 新版本.id
-- 编辑时仅执行 2、3，绝不 UPDATE 旧版本行。
-- =============================================================================
