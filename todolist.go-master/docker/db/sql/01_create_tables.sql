-- Table for tasks
DROP TABLE IF EXISTS `tasks`;

CREATE TABLE `tasks` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `title` varchar(50) NOT NULL,
    `is_done` boolean NOT NULL DEFAULT b'0',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `explanation` varchar(256) NOT NULL DEFAULT "",
    `due_to` date,
    `priority` varchar(10) NOT NULL,  
    `tag` varchar(50),
    `create_user` bigint(20) NOT NULL,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `users`;
 
CREATE TABLE `users` (
    `id`         bigint(20) NOT NULL AUTO_INCREMENT,
    `name`       varchar(50) NOT NULL UNIQUE,
    `password`   binary(32) NOT NULL,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `ownership`;
 
CREATE TABLE `ownership` (
    `user_id` bigint(20) NOT NULL,
    `task_id` bigint(20) NOT NULL,
    PRIMARY KEY (`user_id`, `task_id`)
) DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `tags`;
 
CREATE TABLE `tags` (
    `tag_name` varchar(50) NOT NULL UNIQUE,
    PRIMARY KEY(`tag_name`)
) DEFAULT CHARSET=utf8mb4;

INSERT INTO `tags` (`tag_name`) VALUES ("仕事");
INSERT INTO `tags` (`tag_name`) VALUES ("勉強");
INSERT INTO `tags` (`tag_name`) VALUES ("バイト");
INSERT INTO `tags` (`tag_name`) VALUES ("趣味");
INSERT INTO `tags` (`tag_name`) VALUES ("遊び");
INSERT INTO `tags` (`tag_name`) VALUES ("旅行");
