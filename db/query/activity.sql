select
    {{:by_day  date,}}
    {{:by_day! cast(concat(substring(cast(date as text), 0, 8), '-01') as date) as date,}}
    count(distinct hash) as commits,
    sum(added_sum)       as added,
    sum(removed_sum)     as removed
from commits
where
	repo_id = :repo_id and
	exclude = 0
	{{:author_id and author_id = :author_id}}
	{{:start and date >= :start}}
	{{:end   and date <= :end}}
{{:by_day  group by date}}
{{:by_day! group by substring(cast(date as text), 0, 8 )}}
order by date
