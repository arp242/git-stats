select
	author_id,
	count(*)          as commits,
	sum(added_sum)    as added,
	sum(removed_sum)  as removed,
	min(commits.date) as first,
	max(commits.date) as last,
	names,
	emails
from commits
join authors using (author_id)
where
	commits.repo_id   = :repo_id and
	commits.author_id = :author_id and
	commits.exclude   = 0
	{{:start and commits."date" >= :start}}
	{{:end   and commits."date" <= :end}}
group by author_id, names, emails
order by commits desc, author_id asc
