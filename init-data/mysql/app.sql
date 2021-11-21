CREATE TABLE `s_cc_exchange_symbol_pair`
(
    `exchange_name`  varchar(20) binary NOT NULL DEFAULT '',
    `symbol_pair`    varchar(20) binary NOT NULL DEFAULT '',
    `symbol1`        varchar(20) binary NOT NULL DEFAULT '',
    `symbol2`        varchar(20) binary NOT NULL DEFAULT '',
    `open_timestamp` int       NOT NULL DEFAULT 0,
    `ctime`          timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
    `mtime`          timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY (`exchange_name`, `symbol_pair`)
) ENGINE=InnoDB;

CREATE TABLE `s_cc_exchange_prime_config`
(
    `exchange_name` varchar(20) binary NOT NULL DEFAULT '',
    `symbol_pair`   varchar(20) binary NOT NULL DEFAULT '',
    `status`        varchar(20) binary NOT NULL DEFAULT '',
    `ctime`         timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
    `mtime`         timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY (`exchange_name`, `symbol_pair`)
) ENGINE=InnoDB;
