create table user_master
(
    uid varchar(255) not null,
    username varchar(255) null,
    google_auth_id varchar(100) null,
    google_email varchar(100) null,
    google_profile_image_url varchar(255) null,
    auth_id varchar(50) null,
    auth_encrypted_pw varchar(255) null,
    auth_profile_image_url varchar(255) null,
    constraint user_master_google_auth_id_uindex
        unique (google_auth_id),
    constraint user_master_uid_uindex
        unique (uid)
);

alter table user_master
    add primary key (uid);

create table transactions
(
    txid int auto_increment
        primary key,
    type int not null,
    `from` varchar(255) not null,
    timestamp bigint not null,
    content blob null,
    hash varchar(255) not null,
    constraint transactions_hash_uindex
        unique (hash),
    constraint transactions_user_master_uid_fk
        foreign key (`from`) references user_master (uid)
);

create table blocks
(
    block_hash varchar(255) not null,
    state longblob null,
    block_number bigint null,
    tx_hash varchar(255) null,
    prev_block_hash varchar(255) null,
    constraint blocks_transactions_hash_fk
        foreign key (tx_hash) references transactions (hash)
);

