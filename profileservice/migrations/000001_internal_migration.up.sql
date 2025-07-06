create schema if not exists profile;

create table if not exists profile.profile (
    id bigserial primary key ,
    user_id bigint not null unique ,
    first_name varchar(256) ,
    last_name varchar(256) ,
    created_at timestamp default now(),
    updated_at timestamp,
    bio text
);

create table if not exists profile.follower (
    profile_id bigint not null references profile.profile(id) on delete cascade ,
    follower_id bigint not null references profile.profile(id) on delete cascade,
    followed_at timestamp default now(),
    primary key (profile_id, follower_id),
    check ( profile_id != follower_id )
);

create index on profile.profile(user_id);
create index on profile.follower(profile_id);
create index on profile.follower(follower_id)

