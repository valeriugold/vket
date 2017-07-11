### /* *****************************************************************************
### // Setup the preferences
### // ****************************************************************************/
### SET NAMES utf8 COLLATE 'utf8_unicode_ci';
### SET foreign_key_checks = 1;
### SET time_zone = '-05:00';
### SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
### SET storage_engine = InnoDB;
### SET CHARACTER SET utf8;
### 
### /* *****************************************************************************
### // Remove old database
### // ****************************************************************************/
### DROP DATABASE IF EXISTS vket;
### 
### /* *****************************************************************************
### // Create new database
### // ****************************************************************************/
### CREATE DATABASE vket DEFAULT CHARSET = utf8 COLLATE = utf8_unicode_ci;
### USE vket;

/* *****************************************************************************
// Create the tables
// ****************************************************************************/

drop table if exists event_file;
drop table if exists editor_event;
drop table if exists event;
drop table if exists stored_file;
drop table if exists user;

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

drop table if exists event;
CREATE TABLE event (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    
    name VARCHAR(50) NOT NULL,
    user_id INT(10) UNSIGNED NOT NULL,
    status ENUM('open', 'closed') NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT `f_event_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY (name, user_id),
    PRIMARY KEY (id)
);

drop table if exists stored_file;
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

drop table if exists event_file;
CREATE TABLE event_file (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,

    event_id INT(10) UNSIGNED NOT NULL,
    owner_id INT(10) UNSIGNED NOT NULL,
    /* status ENUM('original', 'processing', 'proposal', 'accepted') NOT NULL, */
    status ENUM('original', 'preview', 'proposal', 'accepted', 'rejected') NOT NULL,
    name VARCHAR(120) NOT NULL,
    stored_file_id INT(10) UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT `f_event_file_event` FOREIGN KEY (`event_id`) REFERENCES `event` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `f_event_file_stored_file` FOREIGN KEY (`stored_file_id`) REFERENCES `stored_file` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `f_event_file_owner` FOREIGN KEY (`owner_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY (name, event_id, owner_id),
    PRIMARY KEY (id)
);

drop table if exists editor_event;
CREATE TABLE editor_event (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,

    editor_id INT(10) UNSIGNED NOT NULL,
    event_id INT(10) UNSIGNED NOT NULL,
    status ENUM('open', 'closed') NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT `f_editor_event_editor` FOREIGN KEY (`editor_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `f_editor_event_event` FOREIGN KEY (`event_id`) REFERENCES `event` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,

    UNIQUE KEY (editor_id, event_id),
    PRIMARY KEY (id)
);

drop table if exists history_user;
CREATE TABLE history_user (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    operation ENUM('insert', 'delete', 'update') NOT NULL,
    user_id INT(10) UNSIGNED NOT NULL,

    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password CHAR(60) NOT NULL,
    role ENUM('admin', 'user', 'editor') NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY (user_id, id),
    UNIQUE KEY (created_at, id),
    /* UNIQUE KEY (email, id), */
    PRIMARY KEY (id)
);

drop table if exists history_event;
CREATE TABLE history_event (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    operation ENUM('insert', 'delete', 'update') NOT NULL,
    event_id INT(10) UNSIGNED NOT NULL,
    
    name VARCHAR(50) NOT NULL,
    user_id INT(10) UNSIGNED NOT NULL,
    status ENUM('open', 'closed') NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY (event_id, id),
    UNIQUE KEY (created_at, id),
    /* UNIQUE KEY (name, id), */
    PRIMARY KEY (id)
);

drop table if exists history_event_file;
CREATE TABLE history_event_file (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    operation ENUM('insert', 'delete', 'update') NOT NULL,
    event_file_id INT(10) UNSIGNED NOT NULL,

    event_id INT(10) UNSIGNED NOT NULL,
    owner_id INT(10) UNSIGNED NOT NULL,
    /* status ENUM('original', 'processing', 'proposal', 'accepted') NOT NULL, */
    status ENUM('original', 'preview', 'proposal', 'accepted', 'rejected') NOT NULL,
    name VARCHAR(120) NOT NULL,
    stored_file_id INT(10) UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY (name, event_id, owner_id, id),
    UNIQUE KEY (event_file_id, id),
    PRIMARY KEY (id)
);

drop table if exists history_stored_file;
CREATE TABLE history_stored_file (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    operation ENUM('insert', 'delete', 'update') NOT NULL,
    stored_file_id INT(10) UNSIGNED NOT NULL,

    name VARCHAR(120) NOT NULL,
    size BIGINT(16) UNSIGNED NOT NULL,
    md5 VARCHAR(32) NOT NULL,
    ref_count INT(10) UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY (md5, id),
    UNIQUE KEY (stored_file_id, id),
    PRIMARY KEY (id)
);

drop table if exists history_editor_event;
CREATE TABLE history_editor_event (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    operation ENUM('insert', 'delete', 'update') NOT NULL,
    editor_event_id INT(10) UNSIGNED NOT NULL,

    editor_id INT(10) UNSIGNED NOT NULL,
    event_id INT(10) UNSIGNED NOT NULL,
    status ENUM('open', 'closed') NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY (editor_id, event_id, id),
    UNIQUE KEY (editor_event_id, event_id),
    PRIMARY KEY (id)
);

/*
history_user
history_event
history_event_file
history_stored_file
history_editor_event

https://stackoverflow.com/questions/12563706/is-there-a-mysql-option-feature-to-track-history-of-changes-to-records
*/

DROP TRIGGER IF EXISTS vket.history_user_insert;
DROP TRIGGER IF EXISTS vket.history_user_update;
DROP TRIGGER IF EXISTS vket.history_user_delete;

DROP TRIGGER IF EXISTS vket.history_event_insert;
DROP TRIGGER IF EXISTS vket.history_event_update;
DROP TRIGGER IF EXISTS vket.history_event_delete;

DROP TRIGGER IF EXISTS vket.history_stored_file_insert;
DROP TRIGGER IF EXISTS vket.history_stored_file_update;
DROP TRIGGER IF EXISTS vket.history_stored_file_delete;

DROP TRIGGER IF EXISTS vket.history_event_file_insert;
DROP TRIGGER IF EXISTS vket.history_event_file_update;
DROP TRIGGER IF EXISTS vket.history_event_file_delete;

DROP TRIGGER IF EXISTS vket.history_editor_event_insert;
DROP TRIGGER IF EXISTS vket.history_editor_event_update;
DROP TRIGGER IF EXISTS vket.history_editor_event_delete;


CREATE TRIGGER vket.history_user_insert AFTER INSERT ON vket.user FOR EACH ROW
    INSERT INTO vket.history_user (operation, user_id, first_name, last_name, email, password, role, created_at) VALUES
    ('insert', NEW.id, NEW.first_name, NEW.last_name, NEW.email, NEW.password, NEW.role, NEW.created_at);
CREATE TRIGGER vket.history_user_update AFTER UPDATE ON vket.user FOR EACH ROW
    INSERT INTO vket.history_user (operation, user_id, first_name, last_name, email, password, role, created_at) VALUES
    ('update', NEW.id, NEW.first_name, NEW.last_name, NEW.email, NEW.password, NEW.role, NEW.updated_at);
CREATE TRIGGER vket.history_user_delete AFTER DELETE ON vket.user FOR EACH ROW
    INSERT INTO vket.history_user (operation, user_id, first_name, last_name, email, password, role) VALUES
    ('delete', OLD.id, OLD.first_name, OLD.last_name, OLD.email, OLD.password, OLD.role);

CREATE TRIGGER vket.history_event_insert AFTER INSERT ON vket.event FOR EACH ROW
    INSERT INTO vket.history_event (operation, event_id, name, user_id, status, created_at) VALUES
    ('insert', NEW.id, NEW.name, NEW.user_id, NEW.status, NEW.created_at);
CREATE TRIGGER vket.history_event_update AFTER UPDATE ON vket.event FOR EACH ROW
    INSERT INTO vket.history_event (operation, event_id, name, user_id, status, created_at) VALUES
    ('update', NEW.id, NEW.name, NEW.user_id, NEW.status, NEW.updated_at);
CREATE TRIGGER vket.history_event_delete AFTER DELETE ON vket.event FOR EACH ROW
    INSERT INTO vket.history_event (operation, event_id, name, user_id, status) VALUES
    ('delete', OLD.id, OLD.name, OLD.user_id, OLD.status);

CREATE TRIGGER vket.history_stored_file_insert AFTER INSERT ON vket.stored_file FOR EACH ROW
    INSERT INTO vket.history_stored_file (operation, stored_file_id, name, size, md5, ref_count, created_at) VALUES
    ('insert', NEW.id, NEW.name, NEW.size, NEW.md5, NEW.ref_count, NEW.created_at);
CREATE TRIGGER vket.history_stored_file_update AFTER UPDATE ON vket.stored_file FOR EACH ROW
    INSERT INTO vket.history_stored_file (operation, stored_file_id, name, size, md5, ref_count, created_at) VALUES
    ('update', NEW.id, NEW.name, NEW.size, NEW.md5, NEW.ref_count, NEW.updated_at);
CREATE TRIGGER vket.history_stored_file_delete AFTER DELETE ON vket.stored_file FOR EACH ROW
    INSERT INTO vket.history_stored_file (operation, stored_file_id, name, size, md5, ref_count) VALUES
    ('delete', OLD.id, OLD.name, OLD.size, OLD.md5, OLD.ref_count);

CREATE TRIGGER vket.history_event_file_insert AFTER INSERT ON vket.event_file FOR EACH ROW
    INSERT INTO vket.history_event_file (operation, event_file_id, event_id, owner_id, status, name, stored_file_id, created_at) VALUES
    ('insert', NEW.id, NEW.event_id, NEW.owner_id, NEW.status, NEW.name, NEW.stored_file_id, NEW.created_at);
CREATE TRIGGER vket.history_event_file_update AFTER UPDATE ON vket.event_file FOR EACH ROW
    INSERT INTO vket.history_event_file (operation, event_file_id, event_id, owner_id, status, name, stored_file_id, created_at) VALUES
    ('update', NEW.id, NEW.event_id, NEW.owner_id, NEW.status, NEW.name, NEW.stored_file_id, NEW.updated_at);
CREATE TRIGGER vket.history_event_file_delete AFTER DELETE ON vket.event_file FOR EACH ROW
    INSERT INTO vket.history_event_file (operation, event_file_id, event_id, owner_id, status, name, stored_file_id) VALUES
    ('delete', OLD.id, OLD.event_id, OLD.owner_id, OLD.status, OLD.name, OLD.stored_file_id);

CREATE TRIGGER vket.history_editor_event_insert AFTER INSERT ON vket.editor_event FOR EACH ROW
    INSERT INTO vket.history_editor_event (operation, editor_event_id, editor_id, event_id, status, created_at) VALUES
    ('insert', NEW.id, NEW.editor_id, NEW.event_id, NEW.status, NEW.created_at);
CREATE TRIGGER vket.history_editor_event_update AFTER UPDATE ON vket.editor_event FOR EACH ROW
    INSERT INTO vket.history_editor_event (operation, editor_event_id, editor_id, event_id, status, created_at) VALUES
    ('update', NEW.id, NEW.editor_id, NEW.event_id, NEW.status, NEW.updated_at);
CREATE TRIGGER vket.history_editor_event_delete AFTER DELETE ON vket.editor_event FOR EACH ROW
    INSERT INTO vket.history_editor_event (operation, editor_event_id, editor_id, event_id, status) VALUES
    ('delete', OLD.id, OLD.editor_id, OLD.event_id, OLD.status);


# add users
INSERT INTO `user` (`first_name`, `last_name`, `email`, `password`, `role`) VALUES
('Azor', 'Popescu', 'aaa@aaa.aaa', 'aaa', 'admin'),
('Grivei', 'Ionescu', 'bbb@aaa.aaa', 'bbb', 'user'),
('a', 'b', 'c@d', 'e', 'user'),
('Labus', 'Georgescu', 'ccc@aaa.aaa', 'ccc', 'editor');

# add events
INSERT INTO `event` (`name`, `user_id`, `status`) VALUES
('birthday 1', (select id from user where email = 'c@d'), 'open'),
('daycare', (select id from user where email = 'c@d'), 'open'),
('preschool', (select id from user where email = 'c@d'), 'open'),
('XbirthdayX', (select id from user where email = 'bbb@aaa.aaa'), 'open'),
('school', (select id from user where email = 'c@d'), 'open'),
('XdaycareX', (select id from user where email = 'bbb@aaa.aaa'), 'open'),
('XpreschoolX', (select id from user where email = 'bbb@aaa.aaa'), 'open'),
('XschoolX', (select id from user where email = 'bbb@aaa.aaa'), 'open');

INSERT INTO editor_event(editor_id, event_id, status) VALUES
(1, (select id from event where name='birthday 1'), 'open'),
(1, (select id from event where name='daycare'), 'open'),
(1, (select id from event where name='preschool'), 'open');

/* create record with id 1, that will be referenced by user_file before the real stored_file is created */
# insert other so called files:
INSERT INTO stored_file (name, size, md5, ref_count) VALUES
("dummy", 0, "", 0xFFFFffff),
("bday1", 100, "abc1", 0),
("bday2", 120, "abc2", 0),
("bday3", 130, "abc3", 0),
("pres1", 100, "abc4", 0),
("pres2", 0,   "abc5", 0),
("pres3", 100, "abc6", 0),
("editedbday1", 100, "abc7", 0),
("editedbday2", 100, "abc8", 0),
("editedbday3", 100, "abc9", 0),
("editedbday4", 100, "ab11", 0);

INSERT INTO event_file(event_id, owner_id, status, name, stored_file_id) VALUES
((select id from event where name='birthday 1'), (select user_id from event where name='birthday 1'), 'original', 'bday1', (select id from stored_file where name='bday1')),
((select id from event where name='birthday 1'), (select user_id from event where name='birthday 1'), 'original', 'bday2', (select id from stored_file where name='bday2')),
((select id from event where name='birthday 1'), (select user_id from event where name='birthday 1'), 'original', 'bday3', (select id from stored_file where name='bday3')),
((select id from event where name='preschool'), (select user_id from event where name='preschool'), 'original', 'pres1', (select id from stored_file where name='pres1')),
((select id from event where name='preschool'), (select user_id from event where name='preschool'), 'original', 'pres2', (select id from stored_file where name='pres2')),
((select id from event where name='preschool'), (select user_id from event where name='preschool'), 'original', 'pres3', (select id from stored_file where name='pres3')),
((select id from event where name='birthday 1'), 1, 'proposal', 'editedbday1', (select id from stored_file where name='editedbday1')),
((select id from event where name='birthday 1'), 1, 'proposal', 'editedbday2', (select id from stored_file where name='editedbday2')),
((select id from event where name='birthday 1'), 1, 'accepted', 'editedbday3', (select id from stored_file where name='editedbday3')),
((select id from event where name='birthday 1'), 1, 'rejected', 'editedbday4', (select id from stored_file where name='editedbday4'));


/*
    size BIGINT(16) UNSIGNED NOT NULL,
    md5 VARCHAR(32) NOT NULL,

UNHEX for BINARY ?
*/
