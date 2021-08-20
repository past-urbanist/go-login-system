create database if not exists user_info;
create table if not exists user_info.user (
    `id` int not null primary key auto_increment,
    `username` varchar(25) not null unique,
    `nickname` varchar(50) default null,
    `password` varchar(32) not null,
    `url` varchar(30) default null
) default charset = utf8;
-- Create a stored procedure: the incoming parameter is the amount of data created
drop procedure if exists user_info.memo_generator;
CREATE PROCEDURE user_info.memo_generator(IN n int) BEGIN
DECLARE i INT DEFAULT 1;
WHILE (i <= n) DO
INSERT INTO user_info.user
VALUES (
        null,
        concat("user", i),
        null,
        md5(left(concat("Password_", i, "qwertYUIOP"), 15)),
        null
    );
SET i = i + 1;
END WHILE;
END;
CALL user_info.memo_generator(10000000);