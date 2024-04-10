// tsort sorts tables.
//
// Use by calling tsort(sel, strcmp), where sel is a list of headers to bind to.
// The default when omitted is 'table.tsort >thead th'
//
// strcmp can be used to pass a string comparison function; the default is
// Intl.Collator().compare, but you can give your own function for a different
// locale, pass any Collator() options, or implement something custom entirely.
//
// Headers with the class "n" will be sorted numerically.
//
// Only rows in the first tbody will be sorted.
//
// data-tsort will be set to "asc" or "desc" for th that's currently sorted.
//
// The initial sort state can be set by adding that to a header. If it's not
// given, tsort-asc will be added to the first header.
//
// https://github.com/arp242/tsort.js | MIT license applies, see LICENSE.
;(function() {
	'use strict';

	window.tsort = function(sel, strcmp) {
		if (sel === undefined)
			sel = 'table.tsort >thead th'
		if (strcmp === undefined)
			strcmp = Intl.Collator().compare

		let initial = false
		document.querySelectorAll(sel).forEach(function(th) {
			initial = initial || th.classList.contains('tsort-asc') || th.classList.contains('tsort-desc')
			th.addEventListener('click', function(e) {
				let num_sort = th.classList.contains('n'),
					col      = Array.from(th.parentNode.children).indexOf(th),
					tbl      = th.closest('table'),
					tbody    = tbl.querySelector('tbody'),
					rows     = Array.from(tbody.children).filter((e) => e.tagName === 'TR'),
					to_i     = (i) => parseFloat(i.replace(/,/g, '')),
					is_sort  = th.classList.contains('tsort-asc')

				if (num_sort)
					rows.sort((a, b) => to_i(b.children[col].innerText) - to_i(a.children[col].innerText))
				else
					rows.sort((a, b) => strcmp(a.children[col].innerText, b.children[col].innerText))

				if (is_sort)
					rows.reverse()

				tbody.innerHTML = ''
				rows.forEach((r) => tbody.appendChild(r))

				tbl.querySelectorAll('th').forEach((t) => t.classList.remove('tsort-asc', 'tsort-desc'))
				th.classList.add(is_sort ? 'tsort-desc' : 'tsort-asc')
			})
		})
		if (!initial)
			document.querySelectorAll('.tsort >thead th')[0].classList.add('tsort-asc')
	}
})();
