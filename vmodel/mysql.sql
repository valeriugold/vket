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

CREATE TABLE stored_file (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,

    name VARCHAR(120) NOT NULL,
    size BIGINT(16) UNSIGNED NOT NULL,
    md5sum BINARY(16) NOT NULL,
    ref_count INT(10) UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY (name),
    UNIQUE KEY (md5sum),
    PRIMARY KEY (id)
);

CREATE TABLE user_file (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,

    name VARCHAR(120) NOT NULL,
    user_id INT(10) UNSIGNED NOT NULL,
    stored_file_id INT(10) UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT `f_user_file_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `f_user_file_stored_file` FOREIGN KEY (`stored_file_id`) REFERENCES `stored_file` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY (name, user_id),
    PRIMARY KEY (id)
);

/*
UNHEX for BINARY ?
*/
