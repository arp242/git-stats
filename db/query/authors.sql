with c as (
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
		commits.repo_id = :repo_id and
		commits.exclude = 0
		{{:start and commits."date" >= :start}}
		{{:end   and commits."date" <= :end}}
	group by author_id, names, emails
	order by commits desc, author_id asc
)
select
	coalesce(cast(c.commits as float) / (select sum(commits) from c) * 100, 0) as commit_perc,
	coalesce(cast(c.added   as float) / (select sum(added)   from c) * 100, 0) as added_perc,
	coalesce(cast(c.removed as float) / (select sum(removed) from c) * 100, 0) as removed_perc,
	c.*
from c
