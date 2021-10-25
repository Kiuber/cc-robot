CREATE TABLE `s_cc_exchange_symbol` (
    `exchange_name` varchar(20) binary NOT NULL DEFAULT '',
    `symbol` varchar(20) binary NOT NULL DEFAULT '',
    `symbol1` varchar(20) binary NOT NULL DEFAULT '',
    `symbol2` varchar(20) binary NOT NULL DEFAULT '',
    `ctime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
    `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY (`exchange_name`, `symbol`)
) ENGINE=InnoDB;
