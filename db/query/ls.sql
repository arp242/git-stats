with x as (
	select
		unnest(files) as file_id,
		count(*)      as num_commits
	from commits
	where repo_id = :repo_id and exclude = 0
	group by file_id
)
select
	x.file_id,
	x.num_commits,
	files.path
from x
join files using (file_id)
order by num_commits desc;

