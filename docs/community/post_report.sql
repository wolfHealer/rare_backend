-- 在 RDS 中执行（本仓库不含 migration，需运维手动建表）
CREATE TABLE IF NOT EXISTS `post_report` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL COMMENT '被举报帖子ID',
  `reporter_id` bigint NOT NULL COMMENT '举报人用户ID',
  `reason_type` varchar(32) NOT NULL COMMENT 'spam/abuse/illegal/misinfo/other',
  `description` varchar(500) DEFAULT NULL COMMENT '补充说明',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '0待处理 1已处理 2已驳回',
  `handled_at` datetime DEFAULT NULL COMMENT '处理时间',
  `handler_id` bigint DEFAULT NULL COMMENT '处理人（管理员）',
  `handle_note` varchar(200) DEFAULT NULL COMMENT '处理备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_post_report_user` (`post_id`, `reporter_id`),
  KEY `idx_post_report_post_id` (`post_id`),
  KEY `idx_post_report_reporter_id` (`reporter_id`),
  KEY `idx_post_report_status` (`status`),
  KEY `idx_post_report_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='帖子举报表';
