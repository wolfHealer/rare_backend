-- 列表/JOIN 复合索引 + 中文关键词 FULLTEXT（ngram）
-- 需 MySQL 8.x；执行前请 make migrate-up

SET NAMES utf8mb4;

-- post 列表：status 过滤 + 排序
ALTER TABLE `post`
  ADD INDEX `idx_post_status_created_at` (`status`, `created_at` DESC),
  ADD INDEX `idx_post_status_type_created_at` (`status`, `type`, `created_at` DESC),
  ADD INDEX `idx_post_status_like_created_at` (`status`, `like_count` DESC, `created_at` DESC);

-- comment 树 / 回复
ALTER TABLE `comment`
  ADD INDEX `idx_comment_target_tree` (`target_type`, `target_id`, `parent_id`, `status`, `created_at`),
  ADD INDEX `idx_comment_root_replies` (`root_id`, `status`, `parent_id`, `created_at`);

-- 批量点赞态：user_id  leading
ALTER TABLE `post_like` ADD INDEX `idx_post_like_user_post` (`user_id`, `post_id`);
ALTER TABLE `post_favorite` ADD INDEX `idx_post_favorite_user_post` (`user_id`, `post_id`);
ALTER TABLE `comment_like` ADD INDEX `idx_comment_like_user_comment` (`user_id`, `comment_id`);

-- JOIN / 筛选
ALTER TABLE `doctor` ADD INDEX `idx_doctor_hospital_audit` (`hospital_id`, `audit_status`);
ALTER TABLE `relief_case`
  ADD INDEX `idx_relief_case_audit_created_at` (`audit_status`, `created_at` DESC),
  ADD INDEX `idx_relief_case_audit_disease` (`audit_status`, `disease_id`);
ALTER TABLE `relief_project` ADD INDEX `idx_relief_project_audit_sort` (`audit_status`, `sort` DESC, `id` DESC);
ALTER TABLE `rehab_institution` ADD INDEX `idx_rehab_audit_region` (`audit_status`, `province_code`, `city_code`);
ALTER TABLE `rehab_institution_disease_rel` ADD INDEX `idx_rehab_disease_institution` (`disease_id`, `institution_id`);
ALTER TABLE `relief_project_disease_rel` ADD INDEX `idx_project_disease_project` (`disease_id`, `project_id`);
ALTER TABLE `drug_relief_project` ADD INDEX `idx_donation_audit_created` (`audit_status`, `created_at` DESC);
ALTER TABLE `drug_channel` ADD INDEX `idx_drug_channel_audit_region` (`audit_status`, `province_code`, `city_code`);
ALTER TABLE `user` ADD INDEX `idx_user_display_name` (`display_name`);

-- FULLTEXT（ngram，适配中文关键词）
ALTER TABLE `post` ADD FULLTEXT INDEX `ft_post_title_content` (`title`, `content`) WITH PARSER ngram;
ALTER TABLE `relief_case` ADD FULLTEXT INDEX `ft_relief_case_search` (`case_title`, `patient_desc`) WITH PARSER ngram;
ALTER TABLE `rare_drug` ADD FULLTEXT INDEX `ft_drug_names` (`generic_name`, `brand_name`) WITH PARSER ngram;
ALTER TABLE `doctor` ADD FULLTEXT INDEX `ft_doctor_search` (`name`, `good_at`) WITH PARSER ngram;
ALTER TABLE `hospital` ADD FULLTEXT INDEX `ft_hospital_search` (`name`, `treat_scope`) WITH PARSER ngram;
ALTER TABLE `rehab_institution` ADD FULLTEXT INDEX `ft_rehab_inst_search` (`name`, `address`) WITH PARSER ngram;
ALTER TABLE `psych_support_org` ADD FULLTEXT INDEX `ft_psych_search` (`name`, `content_intro`) WITH PARSER ngram;
ALTER TABLE `article` ADD FULLTEXT INDEX `ft_article_search` (`title`, `summary`) WITH PARSER ngram;
ALTER TABLE `relief_project` ADD FULLTEXT INDEX `ft_project_search` (`name`, `organizer`) WITH PARSER ngram;
ALTER TABLE `examination_manual` ADD FULLTEXT INDEX `ft_exam_search` (`exam_name`, `exam_purpose`) WITH PARSER ngram;
ALTER TABLE `disease` ADD FULLTEXT INDEX `ft_disease_search` (`name`, `alias`) WITH PARSER ngram;
ALTER TABLE `drug_channel` ADD FULLTEXT INDEX `ft_drug_channel_name` (`name`) WITH PARSER ngram;
ALTER TABLE `help_channel` ADD FULLTEXT INDEX `ft_help_channel_name` (`name`) WITH PARSER ngram;
ALTER TABLE `drug_relief_project` ADD FULLTEXT INDEX `ft_donation_search` (`name`, `organizer`) WITH PARSER ngram;
ALTER TABLE `rehab_train_guide` ADD FULLTEXT INDEX `ft_train_title` (`title`) WITH PARSER ngram;
ALTER TABLE `category` ADD FULLTEXT INDEX `ft_category_name` (`name`) WITH PARSER ngram;
ALTER TABLE `region` ADD FULLTEXT INDEX `ft_region_name` (`name`, `full_name`) WITH PARSER ngram;
ALTER TABLE `user` ADD FULLTEXT INDEX `ft_user_display_name` (`display_name`) WITH PARSER ngram;
