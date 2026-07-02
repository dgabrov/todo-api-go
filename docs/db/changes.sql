-- the changes needed for the database

alter table person add password_salt varchar(64) not null default '';
alter table person add password_payload mediumtext not null default '';