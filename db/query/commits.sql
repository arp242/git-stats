select
	repo_id, hash, date, author_id, added, removed, subject,
	(select array_agg(path) from files where file_id = any(files)
		/* TODO: this needs to fix added/removed as well and exclude=0*/
	) as files
from commits
where
	repo_id   = :repo_id and
	author_id = :author_id and
	exclude   = 0
	{{:start and commits."date" >= :start}}
	{{:end   and commits."date" <= :end}}
order by ctid desc
