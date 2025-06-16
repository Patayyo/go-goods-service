Create Table goods (
    id serial primary key,
    project_id int not null,
    name varchar(255) not null,
    description varchar(255) not null,
    priority int not null,
    removed boolean not null,
    created_at timestamp not null default current_timestamp
);
Create Index idx_goods_project_id ON goods (project_id);
Create Index idx_goods_name ON goods (name);