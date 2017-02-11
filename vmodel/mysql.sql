/* *****************************************************************************
// Setup the preferences
// ****************************************************************************/
SET NAMES utf8 COLLATE 'utf8_unicode_ci';
SET foreign_key_checks = 1;
SET time_zone = '-05:00';
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
SET storage_engine = InnoDB;
SET CHARACTER SET utf8;

/* *****************************************************************************
// Remove old database
// ****************************************************************************/
DROP DATABASE IF EXISTS vket;

/* *****************************************************************************
// Create new database
// ****************************************************************************/
CREATE DATABASE vket DEFAULT CHARSET = utf8 COLLATE = utf8_unicode_ci;
USE vket;

/* *****************************************************************************
// Create the tables
// ****************************************************************************/
CREATE TABLE user (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password CHAR(60) NOT NULL,
    role ENUM('admin', 'user', 'editor') NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY (email),
    
    PRIMARY KEY (id)
);

INSERT INTO `user` (`first_name`, `last_name`, `email`, `password`, `role`) VALUES
('Azor', 'Popescu', 'aaa@aaa.aaa', 'aaa', 'admin'),
('Grivei', 'Ionescu', 'bbb@aaa.aaa', 'bbb', 'user'),
('Labus', 'Georgescu', 'ccc@aaa.aaa', 'ccc', 'editor');

CREATE TABLE event (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    
    name VARCHAR(50) NOT NULL,
    user_id INT(10) UNSIGNED NOT NULL,
    status ENUM('open', 'closed') NOT NULL,

    CONSTRAINT `f_event_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY (name, user_id),
    PRIMARY KEY (id)
);

INSERT INTO `event` (`name`, `user_id`, `status`) VALUES
('birthday 1', (select id from user where email = 'aaa@aaa.aaa'), 'open');

/* drop table if exists stored_file */
/* the actual file name should always be composed of name-md5 */
CREATE TABLE stored_file (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,

    name VARCHAR(120) NOT NULL,
    size BIGINT(16) UNSIGNED NOT NULL,
    md5 VARCHAR(32) NOT NULL,
    ref_count INT(10) UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY (md5),
    PRIMARY KEY (id)
);
/* create record with id 1, that will be referenced by user_file before the real stored_file is created */
INSERT INTO stored_file (name, size, md5, ref_count) VALUES("dummy", 0, 0, 0xFFFFffff);

/* drop table if exists user_file */
CREATE TABLE user_file (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,

    user_id INT(10) UNSIGNED NOT NULL,
    name VARCHAR(120) NOT NULL,
    size BIGINT(16) UNSIGNED NOT NULL,
    md5 VARCHAR(32) NOT NULL,
    stored_file_id INT(10) UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT `f_user_file_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `f_user_file_stored_file` FOREIGN KEY (`stored_file_id`) REFERENCES `stored_file` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY (name, user_id),
    PRIMARY KEY (id)
);

/* drop table if exists user_file */
CREATE TABLE event_file (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,

    event_id INT(10) UNSIGNED NOT NULL,
    name VARCHAR(120) NOT NULL,
    size BIGINT(16) UNSIGNED NOT NULL,
    md5 VARCHAR(32) NOT NULL,
    stored_file_id INT(10) UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT `f_event_file_event` FOREIGN KEY (`event_id`) REFERENCES `event` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `f_event_file_stored_file` FOREIGN KEY (`stored_file_id`) REFERENCES `stored_file` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY (name, event_id),
    PRIMARY KEY (id)
);

/*
UNHEX for BINARY ?
*/
