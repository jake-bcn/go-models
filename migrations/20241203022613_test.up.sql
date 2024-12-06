  CREATE TABLE IF NOT EXISTS `user` (
  `entity_id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `age`  int unsigned not null default 0,
  `created_at` datetime(3),
  `updated_at` datetime(3),
   PRIMARY KEY (`entity_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

 CREATE TABLE IF NOT EXISTS `user_varchar` (
  `entity_id` bigint unsigned NOT NULL,
  `locale` varchar(255),
  `attribute_name` varchar(255),
  `value`  varchar(255),
   CONSTRAINT service_varchar_id_service_service_id FOREIGN KEY (`entity_id`) REFERENCES `user`(`entity_id`) ON DELETE CASCADE ON UPDATE CASCADE,
     UNIQUE KEY service_varchar_entity_id_locale_value (`entity_id`,`locale`,`attribute_name`) 
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;