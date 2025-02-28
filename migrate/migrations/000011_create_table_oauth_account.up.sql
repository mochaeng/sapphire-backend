create table if not exists "oauth_account" (
	provider_id text not null,
	provider_user_id text not null,
	user_id bigint not null,

	constraint oauth_account_provider_id_provider_user_id_pk primary key(provider_id, provider_user_id),
	constraint fk_user_id foreign key (user_id) references "user"(id)
);
