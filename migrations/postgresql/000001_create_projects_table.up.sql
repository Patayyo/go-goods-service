Create Table projects (
    id serial primary key,
    name varchar(255) not null,
    created_at timestamp not null default current_timestamp
);
Insert Into projects (id, name) values (1, 'project start');