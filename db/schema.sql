create table repos (
	repo_id          serial    primary key,
	path             text      not null,
	name             text      not null,
	first_commit     bytea,
	last_commit      bytea,
	first_commit_at  date,
	last_commit_at   date
);
create unique index "repos#path" on repos(lower(path));
create unique index "repos#name" on repos(lower(name));

create table authors (
	author_id        bigserial primary key,
	repo_id          int       not null,
	names            text[]    not null,
	emails           text[]    not null
);
create index "authors#repo_id" on authors(repo_id);
cluster authors using "authors#repo_id";

create table files (
	file_id          bigserial primary key,
	repo_id          int       not null,
	path             text      not null,
	exclude          smallint  not null default 0
);
create index "files#repo_id" on files(repo_id);
cluster files using "files#repo_id";

create table commits (
	repo_id          int       not null,
	hash             bytea     not null,
	date             date      not null,
	author_id        int       not null,
	exclude          smallint  not null default 0,  -- Only commit is a typo fix, or some such.
	files            int[]     not null,  -- 1, 2
	added            int[]     not null,  -- 31, 1 → file[0] added/rm'd/changed 31 lines, file[1] id line
	removed          int[]     not null,
	added_sum        int       not null,  -- Summing the arrays all the time is rather slow
	removed_sum      int       not null,
	subject          text      not null
);
create index "commits#repo_id" on commits(repo_id);
create index "commits#date" on commits(date asc);
cluster commits using "commits#date";

create table events (
	event_id         serial    primary key,
	repo_id          int       not null,
	name             text      not null,
	date             date      not null,
	kind             smallint  not null  -- t=tag, f=fork
);
create index "events#repo_id" on events(repo_id);
cluster events using "events#repo_id";
