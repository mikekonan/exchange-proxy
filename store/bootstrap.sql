create table candles
(
    exchange  TEXT not null,
    pair      TEXT not null,
    timeframe TEXT not null,
    ts        int  not null,
    open      real not null,
    high      real not null,
    low       real not null,
    close     real not null,
    volume    real not null,
    amount    real not null
);

create index candles_exchange_pair_timeframe_ts_index
    on candles (exchange, pair, timeframe, ts);

create index candles_exchange_timeframe_pair_index
    on candles (exchange, timeframe, pair);

