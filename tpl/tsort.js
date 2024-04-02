;(function() {
	'use strict';

	window.tsort = function(sel) {
		if (sel === undefined)
			sel = 'table.tsort th'
		document.querySelectorAll(sel).forEach(function(th) {
			th.addEventListener('click', function(e) {
				let num_sort = th.classList.contains('n'),
					col      = Array.from(th.parentNode.children).indexOf(th),
					tbl      = th.closest('table'),
					tbody    = tbl.querySelector('tbody'),
					rows     = Array.from(tbody.children).filter((e) => e.tagName === 'TR'),
					to_i     = (i) => parseFloat(i.replace(/,/g, '')),
					is_sort  = th.dataset.sort === '1'

				if (num_sort)
					rows.sort((a, b) => to_i(b.children[col].innerText) - to_i(a.children[col].innerText))
				else
					rows.sort((a, b) => a.children[col].innerText.localeCompare(b.children[col].innerText))

				if (is_sort)
					rows.reverse()

				tbody.innerHTML = ''
				rows.forEach((r) => tbody.appendChild(r))

				tbl.querySelectorAll('th').forEach((t) => t.dataset.sort = '0')
				th.dataset.sort = is_sort ? '0' : '1'
			}, false)
		})
	}
})();
