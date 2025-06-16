CREATE TABLE goods_log (
    id Int32,
    project_id Int32,
    name String,
    description String,
    priority Int32,
    removed UInt8,
    event_time DateTime DEFAULT now()
) ENGINE = MergeTree
ORDER BY (id, project_id, name);