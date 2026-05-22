-- MySQL dump 10.13  Distrib 5.7.24, for osx11.1 (x86_64)
--
-- Host: localhost    Database: rare_backend
-- ------------------------------------------------------
-- Server version	8.1.0

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `article`
--

DROP TABLE IF EXISTS `article`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `article` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `title` varchar(255) NOT NULL COMMENT '标题',
  `summary` varchar(1000) DEFAULT '' COMMENT '摘要',
  `cover_image` varchar(500) DEFAULT '' COMMENT '封面',
  `content_type` varchar(50) DEFAULT 'rich_text' COMMENT '内容类型',
  `author_id` bigint DEFAULT NULL,
  `source_name` varchar(255) DEFAULT '' COMMENT '来源',
  `source_url` varchar(500) DEFAULT '' COMMENT '来源链接',
  `status` tinyint DEFAULT '0' COMMENT '0草稿 1待审核 2已发布 3下架',
  `publish_time` datetime DEFAULT NULL,
  `view_count` bigint DEFAULT '0',
  `like_count` bigint DEFAULT '0',
  `favorite_count` bigint DEFAULT '0',
  `is_top` tinyint DEFAULT '0',
  `is_recommend` tinyint DEFAULT '0',
  `seo_title` varchar(255) DEFAULT '',
  `seo_keywords` varchar(500) DEFAULT '',
  `seo_description` varchar(1000) DEFAULT '',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_publish_time` (`publish_time`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `article_block`
--

DROP TABLE IF EXISTS `article_block`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `article_block` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `article_id` bigint NOT NULL,
  `block_type` varchar(50) NOT NULL COMMENT '\ntext\nimage\nvideo\nquote\ntable\nfaq\ndoctor\ndrug\nhospital\ndivider\n',
  `sort_no` int DEFAULT '0',
  `title` varchar(255) DEFAULT '',
  `content` longtext COMMENT '文本内容',
  `extra` json DEFAULT NULL COMMENT '扩展JSON',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_article` (`article_id`),
  KEY `idx_sort` (`article_id`,`sort_no`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `article_tag`
--

DROP TABLE IF EXISTS `article_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `article_tag` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `type` varchar(50) DEFAULT 'general',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `article_tag_rel`
--

DROP TABLE IF EXISTS `article_tag_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `article_tag_rel` (
  `article_id` bigint NOT NULL,
  `tag_id` bigint NOT NULL,
  UNIQUE KEY `uk_rel` (`article_id`,`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `category`
--

DROP TABLE IF EXISTS `category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `parent_id` bigint NOT NULL DEFAULT '0' COMMENT '父分类ID，0为一级',
  `level` tinyint NOT NULL DEFAULT '1' COMMENT '层级：1/2/3',
  `name` varchar(50) NOT NULL COMMENT '分类名称',
  `code` varchar(64) NOT NULL COMMENT '分类编码',
  `description` text COMMENT '分类描述',
  `icon_url` varchar(255) DEFAULT NULL COMMENT '分类图标链接',
  `sort_order` int DEFAULT '0' COMMENT '排序字段',
  `status` tinyint DEFAULT '1' COMMENT '状态：1启用 0停用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`),
  UNIQUE KEY `uk_name_level_parent` (`name`,`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1206 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='疾病分类表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comment`
--

DROP TABLE IF EXISTS `comment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `comment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '评论者用户ID',
  `target_type` varchar(20) NOT NULL COMMENT '评论目标类型：post/knowledge/resource/question',
  `target_id` bigint NOT NULL COMMENT '评论目标ID',
  `parent_id` bigint NOT NULL DEFAULT '0' COMMENT '父评论ID，0为一级评论',
  `root_id` bigint NOT NULL DEFAULT '0' COMMENT '顶级评论ID，0为一级评论',
  `reply_user_id` bigint DEFAULT NULL COMMENT '被回复的用户ID，可为空',
  `content` varchar(500) NOT NULL COMMENT '评论内容',
  `like_count` int NOT NULL DEFAULT '0' COMMENT '评论点赞数',
  `reply_count` int NOT NULL DEFAULT '0' COMMENT '回复数',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1正常 0删除 2审核中 3屏蔽',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_comment_target` (`target_type`,`target_id`),
  KEY `idx_comment_parent_id` (`parent_id`),
  KEY `idx_comment_root_id` (`root_id`),
  KEY `idx_comment_user_id` (`user_id`),
  KEY `idx_comment_reply_user_id` (`reply_user_id`),
  KEY `idx_comment_status` (`status`),
  KEY `idx_comment_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='评论表，支持楼中楼与多模块评论';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `comment_like`
--

DROP TABLE IF EXISTS `comment_like`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `comment_like` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `comment_id` bigint NOT NULL COMMENT '评论ID',
  `user_id` bigint NOT NULL COMMENT '点赞用户ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_comment_user` (`comment_id`,`user_id`),
  KEY `idx_comment_like_user_id` (`user_id`),
  KEY `idx_comment_like_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='评论点赞表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `disease`
--

DROP TABLE IF EXISTS `disease`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `disease` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID，自增',
  `name` varchar(100) NOT NULL COMMENT '病种名称',
  `alias` varchar(200) DEFAULT NULL COMMENT '别名，用逗号分隔',
  `introduction` text COMMENT '病种简介',
  `symptoms` text COMMENT '常见症状',
  `images` json DEFAULT NULL COMMENT '相关图片链接数组',
  `status` tinyint DEFAULT '1' COMMENT '状态：1启用 0停用',
  `creator_id` bigint DEFAULT NULL COMMENT '创建人ID',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='罕见病基础信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `disease_article_rel`
--

DROP TABLE IF EXISTS `disease_article_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `disease_article_rel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `disease_id` bigint NOT NULL,
  `article_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_disease_article` (`disease_id`,`article_id`),
  KEY `idx_article` (`article_id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `disease_category_rel`
--

DROP TABLE IF EXISTS `disease_category_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `disease_category_rel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `disease_id` bigint NOT NULL COMMENT '疾病ID',
  `category_id` bigint NOT NULL COMMENT '分类ID',
  `is_primary` tinyint DEFAULT '0' COMMENT '是否主分类：1是0否',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_disease_category` (`disease_id`,`category_id`),
  KEY `idx_disease_id` (`disease_id`),
  KEY `idx_category_id` (`category_id`)
) ENGINE=InnoDB AUTO_INCREMENT=25 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='疾病-分类关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `disease_tag_rel`
--

DROP TABLE IF EXISTS `disease_tag_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `disease_tag_rel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `disease_id` bigint NOT NULL,
  `tag_id` bigint NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_disease_tag` (`disease_id`,`tag_id`),
  KEY `idx_disease_id` (`disease_id`),
  KEY `idx_tag_id` (`tag_id`)
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='疾病-标签关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor`
--

DROP TABLE IF EXISTS `doctor`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `hospital_id` bigint unsigned NOT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `title` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `department` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `good_at` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `clinic_time` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `contact` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `score` decimal(3,1) DEFAULT '0.0',
  `comment_num` int DEFAULT '0',
  `audit_status` tinyint DEFAULT '0',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_hospital_id` (`hospital_id`),
  KEY `idx_audit_status` (`audit_status`),
  CONSTRAINT `fk_doctors_hospital_id` FOREIGN KEY (`hospital_id`) REFERENCES `hospital` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor_disease_rel`
--

DROP TABLE IF EXISTS `doctor_disease_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor_disease_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `doctor_id` bigint unsigned NOT NULL,
  `disease_id` bigint NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_doctor_disease` (`doctor_id`,`disease_id`),
  KEY `idx_disease_id` (`disease_id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_channel`
--

DROP TABLE IF EXISTS `drug_channel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_channel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `channel_type` varchar(50) NOT NULL,
  `name` varchar(200) NOT NULL,
  `qualification` varchar(500) DEFAULT '',
  `province_code` varchar(255) DEFAULT NULL,
  `city_code` varchar(255) DEFAULT NULL,
  `district_code` varchar(255) DEFAULT NULL,
  `address` varchar(500) DEFAULT '',
  `contact_phone` varchar(50) DEFAULT '',
  `contact_url` varchar(500) DEFAULT '',
  `delivery_scope` varchar(500) NOT NULL,
  `delivery_cycle` varchar(50) NOT NULL,
  `is_insurance_settle` tinyint(1) DEFAULT '0',
  `audit_status` tinyint NOT NULL DEFAULT '0',
  `reject_reason` varchar(255) DEFAULT '',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `province_name` varchar(50) DEFAULT '',
  `city_name` varchar(50) DEFAULT '',
  `district_name` varchar(50) DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_region` (`province_code`,`city_code`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='购药渠道';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_channel_drug_rel`
--

DROP TABLE IF EXISTS `drug_channel_drug_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_channel_drug_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `channel_id` bigint unsigned NOT NULL COMMENT '渠道ID',
  `drug_id` bigint unsigned NOT NULL COMMENT '药品ID',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_channel_drug` (`channel_id`,`drug_id`),
  KEY `idx_channel_id` (`channel_id`),
  KEY `idx_drug_id` (`drug_id`),
  CONSTRAINT `fk_channel_drug_rel_channel` FOREIGN KEY (`channel_id`) REFERENCES `drug_channel` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_channel_drug_rel_drug` FOREIGN KEY (`drug_id`) REFERENCES `rare_drug` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='购药渠道-药品关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_disease_rel`
--

DROP TABLE IF EXISTS `drug_disease_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_disease_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `drug_id` bigint unsigned NOT NULL,
  `disease_id` bigint NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_drug_disease` (`drug_id`,`disease_id`),
  KEY `idx_disease_id` (`disease_id`),
  CONSTRAINT `fk_drug_disease_rel_drug` FOREIGN KEY (`drug_id`) REFERENCES `rare_drug` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='药品疾病关联';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_relief_application`
--

DROP TABLE IF EXISTS `drug_relief_application`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_relief_application` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `project_id` bigint unsigned NOT NULL,
  `user_id` bigint NOT NULL,
  `patient_name` varchar(50) NOT NULL,
  `patient_id_card_enc` varchar(255) NOT NULL,
  `patient_id_card_mask` varchar(32) NOT NULL,
  `diagnosis_proof` varchar(500) NOT NULL,
  `income_proof` varchar(500) NOT NULL,
  `contact_phone` varchar(20) NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'pending',
  `submit_time` datetime NOT NULL,
  `reviewer_id` bigint DEFAULT NULL,
  `review_time` datetime DEFAULT NULL,
  `reject_reason` text,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_user` (`user_id`),
  KEY `idx_status` (`status`),
  CONSTRAINT `fk_relief_application_project` FOREIGN KEY (`project_id`) REFERENCES `drug_relief_project` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='援助申请';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_relief_log`
--

DROP TABLE IF EXISTS `drug_relief_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_relief_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `application_id` bigint unsigned NOT NULL,
  `status` varchar(20) NOT NULL,
  `action_desc` varchar(500) NOT NULL,
  `operator_id` bigint DEFAULT NULL,
  `operator_name` varchar(50) DEFAULT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_application` (`application_id`),
  CONSTRAINT `fk_relief_log_application` FOREIGN KEY (`application_id`) REFERENCES `drug_relief_application` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='援助日志';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_relief_project`
--

DROP TABLE IF EXISTS `drug_relief_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_relief_project` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `drug_id` bigint unsigned NOT NULL,
  `disease_id` bigint NOT NULL,
  `name` varchar(500) NOT NULL,
  `organizer` varchar(200) NOT NULL,
  `apply_condition` text NOT NULL,
  `relief_cycle` varchar(50) NOT NULL,
  `relief_dosage_desc` varchar(100) NOT NULL,
  `apply_form` varchar(500) DEFAULT '',
  `apply_guide` varchar(500) DEFAULT '',
  `material_list` varchar(500) DEFAULT '',
  `progress_query` text NOT NULL,
  `audit_status` tinyint NOT NULL DEFAULT '0',
  `reject_reason` varchar(255) DEFAULT '',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_drug_disease` (`drug_id`,`disease_id`),
  CONSTRAINT `fk_drug_relief_project_drug` FOREIGN KEY (`drug_id`) REFERENCES `rare_drug` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='援助项目';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `exam_manual_disease_rel`
--

DROP TABLE IF EXISTS `exam_manual_disease_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `exam_manual_disease_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `exam_manual_id` bigint unsigned NOT NULL,
  `disease_id` bigint NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exam_manual_disease` (`exam_manual_id`,`disease_id`),
  KEY `idx_disease_id` (`disease_id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `examination_manual`
--

DROP TABLE IF EXISTS `examination_manual`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `examination_manual` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `exam_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `exam_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `exam_purpose` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `reference_value` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `abnormal_interpret` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `sample_notes` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `institution` text COLLATE utf8mb4_unicode_ci,
  `template_excel` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `template_word` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `compare_template` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `audit_status` tinyint DEFAULT '0',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `sort` int DEFAULT '0',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_exam_type` (`exam_type`),
  KEY `idx_audit_status` (`audit_status`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `help_channel`
--

DROP TABLE IF EXISTS `help_channel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `help_channel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `channel_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通道类型：emergency_help/crowdfunding/charity_consulting/foundation_support',
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '求助通道名称',
  `apply_condition` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '求助条件',
  `response_time` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '响应时间',
  `contact_phone` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '联系电话',
  `contact_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '求助入口/官网链接',
  `help_letter_template` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '紧急求助信模板下载地址',
  `crowdfunding_template` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '众筹模板下载地址',
  `audit_status` tinyint NOT NULL DEFAULT '0' COMMENT '审核状态：0待审核 1已通过 2已驳回',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '驳回原因',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序权重',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_channel_type` (`channel_type`),
  KEY `idx_audit_status` (`audit_status`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='求助通道表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hospital`
--

DROP TABLE IF EXISTS `hospital`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `hospital` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `province_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `city_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `district_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `level` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_rare_network` tinyint(1) DEFAULT '0',
  `treat_scope` text COLLATE utf8mb4_unicode_ci COMMENT '文本介绍',
  `address` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `hospital_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `audit_status` tinyint DEFAULT '0' COMMENT '0待审核 1已通过 2已驳回',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `province_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `city_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `district_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `hospital_disease_rel`
--

DROP TABLE IF EXISTS `hospital_disease_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `hospital_disease_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `hospital_id` bigint unsigned NOT NULL,
  `disease_id` bigint NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hospital_disease` (`hospital_id`,`disease_id`),
  KEY `idx_disease_id` (`disease_id`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `medical_insurance_policy`
--

DROP TABLE IF EXISTS `medical_insurance_policy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `medical_insurance_policy` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `disease_id` bigint NOT NULL COMMENT '关联疾病ID',
  `scope_level` tinyint NOT NULL COMMENT '适用范围：1全国 2省级 3市级 4区县级',
  `country_code` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'CN' COMMENT '国家编码，默认CN',
  `province_code` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '省级行政区编码',
  `city_code` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '市级行政区编码',
  `district_code` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '区县编码',
  `province_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '省/直辖市名称',
  `city_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '市名称',
  `district_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '区县名称',
  `policy_title` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '政策标题',
  `policy_original` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '政策原文链接/下载地址',
  `popular_interpret` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '政策通俗解读',
  `reimburse_ratio` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '报销比例说明',
  `reimburse_limit` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '年度报销上限',
  `reimburse_process` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '报销流程图解下载地址',
  `reimburse_material` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '报销材料清单下载地址',
  `remote_apply_template` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '异地备案模板下载地址',
  `publish_date` date DEFAULT NULL COMMENT '政策发布日期',
  `effective_date` date DEFAULT NULL COMMENT '政策生效日期',
  `is_latest` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否最新政策：0否 1是',
  `audit_status` tinyint NOT NULL DEFAULT '0' COMMENT '审核状态：0待审核 1已通过 2已驳回',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '驳回原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_disease_scope_region` (`disease_id`,`scope_level`,`province_code`,`city_code`),
  KEY `idx_is_latest` (`is_latest`),
  KEY `idx_audit_status` (`audit_status`),
  KEY `idx_publish_date` (`publish_date`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='医保政策表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `post`
--

DROP TABLE IF EXISTS `post`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `post` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '发帖用户ID',
  `disease_id` bigint DEFAULT NULL COMMENT '关联病种ID，可为空',
  `category_id` bigint DEFAULT NULL COMMENT '关联疾病分类ID，关联category表，可为空',
  `type` varchar(20) NOT NULL COMMENT '帖子类型：help/experience/emotion/info',
  `title` varchar(100) DEFAULT NULL COMMENT '标题，可为空（移动端发帖可不填）',
  `content` text NOT NULL COMMENT '正文内容',
  `images` json DEFAULT NULL COMMENT '图片URL数组',
  `view_count` int NOT NULL DEFAULT '0' COMMENT '浏览量',
  `like_count` int NOT NULL DEFAULT '0' COMMENT '点赞数',
  `comment_count` int NOT NULL DEFAULT '0' COMMENT '评论数',
  `favorite_count` int NOT NULL DEFAULT '0' COMMENT '收藏数',
  `is_top` tinyint NOT NULL DEFAULT '0' COMMENT '是否置顶：0否 1是',
  `is_recommend` tinyint NOT NULL DEFAULT '0' COMMENT '是否推荐：0否 1是',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '状态：0审核中 1正常 2驳回 3删除',
  `reject_reason` varchar(200) DEFAULT NULL COMMENT '驳回原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_post_user_id` (`user_id`),
  KEY `idx_post_disease_id` (`disease_id`),
  KEY `idx_post_category_id` (`category_id`),
  KEY `idx_post_type` (`type`),
  KEY `idx_post_status` (`status`),
  KEY `idx_post_created_at` (`created_at`),
  KEY `idx_post_top_recommend` (`is_top`,`is_recommend`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='社区帖子表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `post_favorite`
--

DROP TABLE IF EXISTS `post_favorite`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `post_favorite` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL COMMENT '帖子ID',
  `user_id` bigint NOT NULL COMMENT '收藏用户ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_post_favorite_user` (`post_id`,`user_id`),
  KEY `idx_post_favorite_user_id` (`user_id`),
  KEY `idx_post_favorite_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='帖子收藏表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `post_like`
--

DROP TABLE IF EXISTS `post_like`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `post_like` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL COMMENT '帖子ID',
  `user_id` bigint NOT NULL COMMENT '点赞用户ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_post_user` (`post_id`,`user_id`),
  KEY `idx_post_like_user_id` (`user_id`),
  KEY `idx_post_like_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='帖子点赞表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `psych_support_org`
--

DROP TABLE IF EXISTS `psych_support_org`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `psych_support_org` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '心理支持机构名称',
  `province_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `city_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `district_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `address` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '机构地址',
  `contact_phone` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '联系电话',
  `contact_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '官网/预约入口',
  `is_free` tinyint(1) DEFAULT NULL COMMENT '是否免费：0否 1是',
  `consult_way` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '咨询方式：online/offline/both',
  `content_intro` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '机构简介/服务说明',
  `audit_status` tinyint NOT NULL DEFAULT '0' COMMENT '审核状态：0待审核 1已通过 2已驳回',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '驳回原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `province_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `city_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `district_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_region` (`province_code`,`city_code`),
  KEY `idx_audit_status` (`audit_status`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='心理支持机构表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `psych_support_org_disease_rel`
--

DROP TABLE IF EXISTS `psych_support_org_disease_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `psych_support_org_disease_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `org_id` bigint unsigned NOT NULL COMMENT '心理机构ID',
  `disease_id` bigint NOT NULL COMMENT '疾病ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_psych_support_org_disease` (`org_id`,`disease_id`),
  KEY `idx_disease_id` (`disease_id`),
  CONSTRAINT `fk_psych_support_org_disease_rel_org_id` FOREIGN KEY (`org_id`) REFERENCES `psych_support_org` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='心理支持机构与疾病关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `psych_support_resource`
--

DROP TABLE IF EXISTS `psych_support_resource`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `psych_support_resource` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `disease_id` bigint NOT NULL COMMENT '关联疾病ID',
  `resource_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源类型：guide/manual/hotline/online_resource',
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源名称',
  `content_intro` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源内容简介',
  `guide_pdf` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '心理疏导指南PDF下载地址',
  `manual_patient` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '患者心理调节手册下载地址',
  `manual_family` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '家属心理支持手册下载地址',
  `contact_phone` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '联系电话/热线',
  `contact_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '官网/线上入口链接',
  `audit_status` tinyint NOT NULL DEFAULT '0' COMMENT '审核状态：0待审核 1已通过 2已驳回',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '驳回原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_disease_type` (`disease_id`,`resource_type`),
  KEY `idx_audit_status` (`audit_status`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='心理支持资源表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rare_drug`
--

DROP TABLE IF EXISTS `rare_drug`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rare_drug` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `generic_name` varchar(200) NOT NULL COMMENT '通用名',
  `brand_name` varchar(200) DEFAULT '' COMMENT '商品名',
  `indication` text NOT NULL COMMENT '适应症',
  `dosage_form` varchar(50) NOT NULL COMMENT '剂型',
  `spec` varchar(100) NOT NULL COMMENT '规格',
  `ref_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '参考价格',
  `drug_type` varchar(30) NOT NULL COMMENT 'origin_import/origin_domestic/generic/other',
  `is_insurance` tinyint(1) NOT NULL DEFAULT '0',
  `has_relief` tinyint(1) NOT NULL DEFAULT '0',
  `is_launched` tinyint(1) NOT NULL DEFAULT '0',
  `need_prescription` tinyint(1) NOT NULL DEFAULT '1',
  `manual_original` varchar(500) DEFAULT '',
  `manual_popular` varchar(500) DEFAULT '',
  `audit_status` tinyint NOT NULL DEFAULT '0' COMMENT '0待审核 1通过 2驳回',
  `reject_reason` varchar(255) DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_generic_name` (`generic_name`),
  KEY `idx_brand_name` (`brand_name`),
  KEY `idx_audit_status` (`audit_status`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='药品表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `region`
--

DROP TABLE IF EXISTS `region`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `region` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(20) NOT NULL COMMENT '行政区编码（如110000）',
  `name` varchar(50) NOT NULL COMMENT '名称',
  `full_name` varchar(100) NOT NULL COMMENT '全称（如北京市）',
  `parent_code` varchar(20) DEFAULT '' COMMENT '父级编码',
  `level` tinyint NOT NULL COMMENT '1省 2市 3区县',
  `sort` int DEFAULT '0' COMMENT '排序',
  `is_enabled` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_parent` (`parent_code`),
  KEY `idx_level` (`level`)
) ENGINE=InnoDB AUTO_INCREMENT=137 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='行政区划表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rehab_institution`
--

DROP TABLE IF EXISTS `rehab_institution`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rehab_institution` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '康复机构名称',
  `province_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `city_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `district_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `qualification` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '机构资质证明地址',
  `rehab_projects` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '康复项目范围',
  `fee_standard` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '收费标准',
  `contact_phone` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '咨询电话',
  `contact_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '官网/预约链接',
  `address` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '机构地址',
  `audit_status` tinyint NOT NULL DEFAULT '0' COMMENT '审核状态：0待审核 1已通过 2已驳回',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '驳回原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `province_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `city_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `district_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_region` (`province_code`,`city_code`),
  KEY `idx_audit_status` (`audit_status`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='康复机构表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rehab_institution_disease_rel`
--

DROP TABLE IF EXISTS `rehab_institution_disease_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rehab_institution_disease_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `institution_id` bigint unsigned NOT NULL COMMENT '康复机构ID',
  `disease_id` bigint NOT NULL COMMENT '疾病ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rehab_institution_disease` (`institution_id`,`disease_id`),
  KEY `idx_disease_id` (`disease_id`),
  CONSTRAINT `fk_rehab_institution_disease_rel_institution_id` FOREIGN KEY (`institution_id`) REFERENCES `rehab_institution` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='康复机构与疾病关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rehab_train_guide`
--

DROP TABLE IF EXISTS `rehab_train_guide`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rehab_train_guide` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `rehab_stage` varchar(50) NOT NULL COMMENT '阶段：early/middle/late/stable/progressive',
  `title` varchar(500) NOT NULL COMMENT '指南标题',
  `train_purpose` varchar(500) NOT NULL COMMENT '训练目的',
  `train_content` text NOT NULL COMMENT '训练内容',
  `forbidden_action` text NOT NULL COMMENT '禁忌动作',
  `pic_urls` json DEFAULT NULL COMMENT '图片数组',
  `guide_pdf` varchar(500) DEFAULT '' COMMENT 'PDF下载',
  `guide_word` varchar(500) DEFAULT '' COMMENT 'Word下载',
  `audit_status` tinyint DEFAULT '0' COMMENT '0待审核 1通过 2驳回',
  `reject_reason` varchar(255) DEFAULT '',
  `sort` int DEFAULT '0',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_stage` (`rehab_stage`),
  KEY `idx_audit` (`audit_status`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='康复训练指南';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rehab_train_guide_disease_rel`
--

DROP TABLE IF EXISTS `rehab_train_guide_disease_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rehab_train_guide_disease_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `guide_id` bigint unsigned NOT NULL,
  `disease_id` bigint NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_guide_disease` (`guide_id`,`disease_id`),
  KEY `idx_disease` (`disease_id`),
  CONSTRAINT `fk_rehab_guide` FOREIGN KEY (`guide_id`) REFERENCES `rehab_train_guide` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='康复指南-疾病关联';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `relief_case`
--

DROP TABLE IF EXISTS `relief_case`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `relief_case` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `disease_id` bigint NOT NULL COMMENT '关联疾病ID',
  `project_id` bigint unsigned NOT NULL COMMENT '关联救助项目ID',
  `case_title` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '案例标题',
  `patient_desc` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '匿名化患者病情/家庭情况简介',
  `apply_cycle` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '申请周期',
  `actual_relief` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '实际救助金额/福利',
  `experience` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '申请经验分享',
  `pitfall_guide` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '申请避坑要点',
  `case_pdf` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '案例PDF下载地址',
  `material_template` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '成功案例材料模板下载地址',
  `audit_status` tinyint NOT NULL DEFAULT '0' COMMENT '审核状态：0待审核 1已通过 2已驳回',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '驳回原因',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_disease_project` (`disease_id`,`project_id`),
  KEY `idx_audit_status` (`audit_status`),
  KEY `fk_relief_case_project_id` (`project_id`),
  CONSTRAINT `fk_relief_case_project_id` FOREIGN KEY (`project_id`) REFERENCES `relief_project` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='救助案例表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `relief_project`
--

DROP TABLE IF EXISTS `relief_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `relief_project` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `relief_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '救助类型：medical_cost/living_support/rehab_subsidy/drug_relief/other',
  `name` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '救助项目名称',
  `organizer` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '主办机构',
  `apply_condition` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '申请条件',
  `relief_standard` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '救助金额/标准',
  `apply_deadline` date DEFAULT NULL COMMENT '申请截止时间，长期有效可为空',
  `apply_difficulty` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '申请难度：easy/medium/hard',
  `apply_process` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '申请流程',
  `apply_form` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '申请表下载地址',
  `apply_guide` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '申请指南下载地址',
  `material_list` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '材料清单下载地址',
  `contact_phone` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '咨询电话',
  `contact_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '咨询链接/官网链接',
  `audit_status` tinyint NOT NULL DEFAULT '0' COMMENT '审核状态：0待审核 1已通过 2已驳回',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '驳回原因',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序权重',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_relief_type` (`relief_type`),
  KEY `idx_apply_deadline` (`apply_deadline`),
  KEY `idx_audit_status` (`audit_status`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公益救助项目表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `relief_project_disease_rel`
--

DROP TABLE IF EXISTS `relief_project_disease_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `relief_project_disease_rel` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` bigint unsigned NOT NULL COMMENT '救助项目ID',
  `disease_id` bigint NOT NULL COMMENT '疾病ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_disease` (`project_id`,`disease_id`),
  KEY `idx_disease_id` (`disease_id`),
  CONSTRAINT `fk_relief_project_disease_rel_project_id` FOREIGN KEY (`project_id`) REFERENCES `relief_project` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=44 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='救助项目与疾病关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tag`
--

DROP TABLE IF EXISTS `tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tag` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '标签名',
  `code` varchar(64) NOT NULL COMMENT '标签编码',
  `sort_order` int DEFAULT '0',
  `status` tinyint DEFAULT '1',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='标签表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user`
--

DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `phone` varchar(20) NOT NULL COMMENT '手机号登录',
  `password_hash` varchar(255) NOT NULL COMMENT '密码哈希',
  `display_name` varchar(50) DEFAULT NULL COMMENT '展示名（半匿名）',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像URL',
  `role` tinyint NOT NULL DEFAULT '1' COMMENT '1普通用户 2专家 9管理员',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '1正常 0禁用',
  `last_login_at` datetime DEFAULT NULL,
  `login_count` int NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `phone` (`phone`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_follow`
--

DROP TABLE IF EXISTS `user_follow`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_follow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '关注发起人',
  `follow_user_id` bigint NOT NULL COMMENT '被关注用户',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_follow` (`user_id`,`follow_user_id`),
  KEY `idx_follow_user_id` (`follow_user_id`),
  KEY `idx_user_follow_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户关注关系表';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-05-21  1:33:40




-- 创建短信验证码表
CREATE TABLE IF NOT EXISTS sms_codes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(10) NOT NULL,
    scene VARCHAR(20) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expired_at DATETIME NOT NULL,
    used TINYINT(1) NOT NULL DEFAULT 0,
    updated_at DATETIME NULL ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_phone_scene (phone, scene),
    INDEX idx_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;