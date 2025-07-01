create table person
(
    person_id   varchar(64)  not null,
    provided_id varchar(128) not null,
    person_name varchar(255) not null,
    login       varchar(64)  not null
);

create table session
(
    session_id  varchar(64) not null,
    updated_dt  datetime    not null,
    token       varchar(64) not null,
    expired_ind varchar(1)  not null
);

create table session_person
(
    session_person_id varchar(64) not null,
    session_id        varchar(64) not null,
    person_id         varchar(64) not null
);

create table todo_item
(
    todo_item_id  varchar(64) not null,
    person_id     varchar(64) not null,
    comments      mediumtext  not null,
    project_cd    varchar(255),
    context_cd    varchar(255),
    priority      int(11) not null,
    added_dt      datetime    not null,
    due_dt        datetime,
    completed_ind varchar(1),
    updated_dt datetime not null
);

-- primary keys
alter table person
    add primary key pk_person (person_id);
alter table session
    add primary key pk_session (session_id);
alter table session_person
    add primary key pk_session_person (session_person_id);
alter table todo_item
    add primary key pk_todo_item (todo_item_id);

-- foreign keys
alter table session_person
    add constraint fk_sp_session foreign key (session_id) references session (session_id);

alter table session_person
    add constraint fk_sp_person foreign key (person_id) references person (person_id);

alter table todo_item
    add constraint fk_todo_item_person foreign key (person_id) references person (person_id);


-- only one user with provided_id
alter table person
    add constraint ux_provided_id unique (provided_id);

-- adding tables for the attach manager

CREATE TABLE `item` (
                        `item_id` VARCHAR (192) NOT NULL,
                        `person_id` VARCHAR (192),
                        `item_name` VARCHAR (765),
                        `description` mediumtext ,
                        `category` VARCHAR (765),
                        `flag_ind` VARCHAR (3),
                        `added_dt` DATETIME ,
                        `updated_dt` DATETIME
);

CREATE TABLE `attachment` (
                              `attachment_id` VARCHAR (192) NOT NULL,
                              `item_id` VARCHAR (192) NOT NULL,
                              `description` VARCHAR (765),
                              `file_name` VARCHAR (192),
                              `content_type` VARCHAR (384),
                              `seq_no` BIGINT (20),
                              `added_dt` datetime ,
                              `updated_dt` datetime
);

ALTER TABLE item ADD CONSTRAINT PRIMARY KEY(item_id);
ALTER TABLE item ADD CONSTRAINT fk_item_person FOREIGN KEY (person_id) REFERENCES person(person_id);

ALTER TABLE attachment ADD CONSTRAINT PRIMARY KEY(attachment_id);
ALTER TABLE attachment ADD CONSTRAINT fk_att_item FOREIGN KEY (item_id) REFERENCES item(item_id);

-- added