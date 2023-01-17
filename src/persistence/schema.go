package persistence

var SchemaUp = []string{
	`CREATE TABLE deposit_histories (
	    ts TIMESTAMP,
	    amount DECIMAL(20,8)
	);`,
	`CREATE TABLE deposit_hourly (
	    ts TIMESTAMP PRIMARY KEY,
	    amount DECIMAL(20,8)
	);`,
}
