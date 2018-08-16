DROP DATABASE IF EXISTS `UserDB`;
CREATE DATABASE `UserDB`;
USE `UserDB`;
 
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` bigint(64) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) NOT NULL,
  `realname` varchar(1024) DEFAULT NULL UNIQUE,
  `nickname` varchar(1024) DEFAULT NULL,
  `pwd` varchar(32) DEFAULT NULL,
  `role` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0 normal user , 1 manager',
  `more` text COMMENT 'redundant field',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `uuidIdx` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS `contact`;
CREATE TABLE `contact` (
  `id` bigint(64) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) NOT NULL COMMENT 'uuid',
  `addressee` text,
  `address` text,
  `telephone` text,
  `more` text COMMENT 'redundant field',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `uuidIdx` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS `login`;
CREATE TABLE `login` (
  `id` bigint(64) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) NOT NULL COMMENT 'uuid',
  `ip` varchar(20) DEFAULT NULL,
  `GPS` varchar(255) DEFAULT NULL,
  `device` varchar(255) DEFAULT NULL,
  `browser` varchar(255) DEFAULT NULL,
  `status` int(11) NOT NULL COMMENT '0 successful login , 1 otherwise ',
  `more` text COMMENT 'redundant field',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `uuidIdx` (`uuid`),
  KEY `statusIdx` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS `avatar`;
CREATE TABLE `avatar` (
  `id` bigint(64) NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) NOT NULL,
  `pid` varchar(128) NOT NULL COMMENT 'photo id',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `uuidIdx` (`uuid`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;



