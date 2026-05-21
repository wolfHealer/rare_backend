-- Rollback baseline: drop all application tables
-- WARNING: destroys all data. Use in dev/test only.

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `user_follow`;
DROP TABLE IF EXISTS `user`;
DROP TABLE IF EXISTS `tag`;
DROP TABLE IF EXISTS `relief_project_disease_rel`;
DROP TABLE IF EXISTS `relief_project`;
DROP TABLE IF EXISTS `relief_case`;
DROP TABLE IF EXISTS `rehab_train_guide_disease_rel`;
DROP TABLE IF EXISTS `rehab_train_guide`;
DROP TABLE IF EXISTS `rehab_institution_disease_rel`;
DROP TABLE IF EXISTS `rehab_institution`;
DROP TABLE IF EXISTS `region`;
DROP TABLE IF EXISTS `rare_drug`;
DROP TABLE IF EXISTS `psych_support_resource`;
DROP TABLE IF EXISTS `psych_support_org_disease_rel`;
DROP TABLE IF EXISTS `psych_support_org`;
DROP TABLE IF EXISTS `post_like`;
DROP TABLE IF EXISTS `post_favorite`;
DROP TABLE IF EXISTS `post`;
DROP TABLE IF EXISTS `medical_insurance_policy`;
DROP TABLE IF EXISTS `hospital_disease_rel`;
DROP TABLE IF EXISTS `hospital`;
DROP TABLE IF EXISTS `help_channel`;
DROP TABLE IF EXISTS `examination_manual`;
DROP TABLE IF EXISTS `exam_manual_disease_rel`;
DROP TABLE IF EXISTS `drug_relief_project`;
DROP TABLE IF EXISTS `drug_relief_log`;
DROP TABLE IF EXISTS `drug_relief_application`;
DROP TABLE IF EXISTS `drug_disease_rel`;
DROP TABLE IF EXISTS `drug_channel_drug_rel`;
DROP TABLE IF EXISTS `drug_channel`;
DROP TABLE IF EXISTS `doctor_disease_rel`;
DROP TABLE IF EXISTS `doctor`;
DROP TABLE IF EXISTS `disease_tag_rel`;
DROP TABLE IF EXISTS `disease_category_rel`;
DROP TABLE IF EXISTS `disease_article_rel`;
DROP TABLE IF EXISTS `disease`;
DROP TABLE IF EXISTS `comment_like`;
DROP TABLE IF EXISTS `comment`;
DROP TABLE IF EXISTS `category`;
DROP TABLE IF EXISTS `article_tag_rel`;
DROP TABLE IF EXISTS `article_tag`;
DROP TABLE IF EXISTS `article_block`;
DROP TABLE IF EXISTS `article`;

SET FOREIGN_KEY_CHECKS = 1;
