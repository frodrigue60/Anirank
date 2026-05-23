package testutil

// SQLIPayloads es una lista comprensiva de vectores de ataque SQLi para testing
var SQLIPayloads = []string{
	`' OR '1'='1`,
	`' OR 1=1 --`,
	`admin' --`,
	`' UNION SELECT NULL, NULL, NULL --`,
	`' UNION SELECT username, password FROM users --`,
	`'; DROP TABLE users; --`,
	`' OR pg_sleep(2) --`,
	`'; SELECT pg_sleep(2) --`,
	`') AND CAST(version() AS INT)=1 --`,
	`1; SELECT pg_sleep(2)`,
}
