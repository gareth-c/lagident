CREATE TABLE technologies (
  name    VARCHAR(255),
  details VARCHAR(255)
);
insert into technologies values (
  'Go', 'An open source programming language that makes it easy to build simple and efficient software.'
);
insert into technologies values (
  'JavaScript', 'A lightweight, interpreted, or just-in-time compiled programming language with first-class functions.'
);
insert into technologies values (
  'MySQL', 'A powerful, open source object-relational database'
);


CREATE TABLE IF NOT EXISTS `targets` (
    `uuid`       CHAR(36) NOT NULL PRIMARY KEY,
    `name`       VARCHAR(255) NOT NULL,
    `address`    VARCHAR(255) NOT NULL
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8
  COLLATE = utf8_general_ci;

INSERT INTO `targets` VALUES (
  '38c84db2-1c79-40c6-86aa-650474f2cc88', 'localhost', '127.0.0.1'
);

CREATE TABLE IF NOT EXISTS `statistics` (
    `target_uuid` CHAR(36) NOT NULL PRIMARY KEY,
    `state`       VARCHAR(10) NOT NULL,
    `sent`        BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `recv`        BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `last`        DOUBLE DEFAULT 0,
    `loss`        DOUBLE DEFAULT 0,
    `sum`         DOUBLE DEFAULT 0,
    `max`         DOUBLE DEFAULT 0,
    `min`         DOUBLE DEFAULT NULL,
    `avg15m`      DOUBLE DEFAULT 0,
    `avg6h`       DOUBLE DEFAULT 0,
    `avg24h`      DOUBLE DEFAULT 0,
    `timestamp`   BIGINT(20) NOT NULL
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8
  COLLATE = utf8_general_ci;

CREATE TABLE IF NOT EXISTS `losses` (
    `target_uuid` CHAR(36) NOT NULL,
    `timestamp`   BIGINT(20) NOT NULL,
    PRIMARY KEY (`target_uuid`, `timestamp`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8
  COLLATE = utf8_general_ci
  COMMENT =  "When a target is unreachable, a record is inserted into this table";

CREATE TABLE IF NOT EXISTS `latencies` (
    `target_uuid` CHAR(36) NOT NULL,
    `timestamp`   BIGINT(20) NOT NULL,
    `latency`     DOUBLE NOT NULL,
    PRIMARY KEY (`target_uuid`, `timestamp`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8
  COLLATE = utf8_general_ci
  COMMENT =  "Time Series data of the latency per target";

CREATE TABLE IF NOT EXISTS `histograms` (
    `target_uuid` CHAR(36) NOT NULL,
    `timestamp`   BIGINT(20) NOT NULL,
    `bucket`      DOUBLE DEFAULT 0,
    `count`       INTEGER UNSIGNED DEFAULT 1,
    PRIMARY KEY (`target_uuid`, `timestamp`, `bucket`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8
  COLLATE = utf8_general_ci;