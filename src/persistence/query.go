package persistence

const insertHistoriesQuery = `
  INSERT INTO deposit_histories (ts, amount)
    VALUES ($1, $2)`

const updatePostHourlyQuery = `
  UPDATE deposit_hourly
    SET amount = amount + $2
    WHERE ts > date_trunc('hour', $1::timestamp - interval '1 second') + interval '1 hour'`

const insertHourlyQuery = `
  INSERT INTO deposit_hourly (ts, amount)
    SELECT
      date_trunc('hour', $1::timestamp - interval '1 second') + interval '1 hour' AS ts,
      COALESCE((
        SELECT amount + $2
          FROM deposit_hourly
          WHERE ts < date_trunc('hour', $1::timestamp - interval '1 second') + interval '1 hour'
          ORDER BY ts DESC LIMIT 1
        ), $2)
    ON CONFLICT (ts) DO UPDATE SET amount = deposit_hourly.amount + $2`

const selectHourlyQuery = `
  (
    SELECT date_trunc('hour', $1::timestamp - interval '1 second') + interval '1 hour' AS ts, amount
	FROM deposit_hourly
	WHERE ts <= date_trunc('hour', $1::timestamp - interval '1 second') + interval '1 hour'
	ORDER BY ts DESC LIMIT 1
  ) UNION (
    SELECT ts, amount
    FROM deposit_hourly
    WHERE ts > date_trunc('hour', $1::timestamp - interval '1 second') + interval '1 hour'
      AND ts <= $2
  ) ORDER BY ts`
