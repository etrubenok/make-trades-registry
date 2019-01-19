CREATE TYPE maketrades.symbol_info(symbol text,
    status text,
    asset text,
    asset_precision bigint,
    quote text,
    quote_precision bigint,
    order_types list<text>,
    iceberg_allowed int);

CREATE TABLE maketrades.symbols_snapshots(year int,
    month int,
    day int,
    market_id int,
    snapshot_time timestamp,
    symbols list<FROZEN<maketrades.symbol_info>>,
    PRIMARY KEY ((year, month, day, market_id), snapshot_time)
)
WITH CLUSTERING ORDER BY (snapshot_time DESC);