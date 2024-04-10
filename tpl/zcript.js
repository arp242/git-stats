let style = (name) => getComputedStyle(document.documentElement).getPropertyValue(`--${name}`)

// Create Date() object from "year-month-day hour:min:sec" string. Any of the
// parts to the right may be missing: "2017-06" will create a date on June 1st.
let get_date = function(str) {
	let s = str.split(/[: TZ-]/)
	return new Date(s[0],
		parseInt((s[1] || 1), 10) - 1,
		(s[2] || 1),
		(s[3] || 0), (s[4] || 0), (s[5] || 0))
}

// Format a number with a thousands separator. https://stackoverflow.com/a/2901298/660921
var format_int = (n) => (n+'').replace(/\B(?=(\d{3})+(?!\d))/g, String.fromCharCode(USER_SETTINGS.number_format))

var months      = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'],
	days        = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],
	monthsShort = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
	daysShort   = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

let USER_SETTINGS = {date_format: '2 Jan 2006'}

// Format a date according to user configuration.
let format_date = function(date, plusDays) {
	if (typeof(date) === 'string')
		date = get_date(date)

	if (plusDays)
		date = new Date(date.setDate(date.getDate() + plusDays))

	let m        = date.getMonth() + 1,
		d        = date.getDate(),
		items    = USER_SETTINGS.date_format.split(/[-/\s]/),
		new_date = []

	// Simple implementation of Go's time format. We only offer the current
	// formatters, so that's all we support:
	//   "2006-01-02", "02-01-2006", "01/02/06", "2 Jan 06", "Mon Jan 2 2006"
	for (let i = 0; i < items.length; i++) {
		switch (items[i]) {
			case '2006': new_date.push(date.getFullYear());                  break;
			case '06':   new_date.push((date.getFullYear() + '').substr(2)); break;
			case '01':   new_date.push(m >= 10 ? m : ('0' + m));             break;
			case '02':   new_date.push(d >= 10 ? d : ('0' + d));             break;
			case '2':    new_date.push(d);                                   break;
			case 'Jan':  new_date.push(monthsShort[date.getMonth()]);        break;
			case 'Mon':  new_date.push(daysShort[date.getDay()]);            break;
		}
	}

	let joiner = '-'
	if (USER_SETTINGS.date_format.indexOf('/') > -1)
		joiner = '/'
	else if (USER_SETTINGS.date_format.indexOf(' ') > -1)
		joiner = ' '
	return new_date.join(joiner)
}

let esc = (s) => new Option(s).innerHTML

let kind_name = (k) => {
	switch (String.fromCharCode(e.kind)) {
		case 't': return 'tag'
		case 'f': return 'fork'
		case 'l': return 'license change'
		case 'o': return 'owner change'
		default:  return 'other'
	}
}

let draw_chart = () => {
	let canvas = window.graph
	if (!canvas || canvas.dataset.done === 't')
		return
	canvas.dataset.done = 't'
	let ctx = canvas.getContext('2d', {alpha: false})

	let stats = JSON.parse(canvas.dataset.stats)
	if (!stats)
		return

	let events = {}
	for (e of JSON.parse(canvas.dataset.events)) {
		if (String.fromCharCode(e.kind) === 't')
			continue
		events[e.date] = `${kind_name(e.kind)}: ${e.name}`
	}

	// Group charts by week; just the daily stats are too noisy IMO.
	//
	// TODO should probably do this in SQL, and anchor on monday or something.
	// But this is quick and easy for now.
	let weekly = stats.length > 185
	if (weekly) {
		let c    = 0,
			ev   = {},
			last = ''
		stats.forEach((s, i) => {
			if (i % 7 === 0) {
				stats[i].commits = c
				last = s.date
				if (events[s.date])
					ev[s.date] = events[s.date]
				c = 0
			} else {
				c += s.commits
				if (events[s.date])
					ev[last] = events[s.date]
			}
		})
		stats = stats.filter((_, i) => i % 7 === 0)
		events = ev
	}

	// It's not too uncommon to have an enormously large amount of commits for
	// just one or two days for various reasons. It's okay if these are
	// (literally) off the charts because it really throws things off otherwise.
	let data = stats.map((s) => s.commits),
		max  = 0

	if (weekly) {
		let sd   = data.toSorted((a, b) => b - a),
			prev = sd[10]
		for (let n of sd.slice(0, 10).reverse()) {
			if (n > prev * 2)
				break;
			max = n
		}
		//console.log(max, '→', sd.slice(0, 10))
	} else
		data.forEach((n) => max = Math.max(max, n))

	window.max.innerHTML = `${max}`
	window.half.innerHTML = `${Math.round(max / 2)}`

	let chart = charty(ctx, data, {
		grid: [2.5, 50, 97.5],
		max:  max,
		line: {
			color: style('chart-line'),
			fill:  style('chart-fill'),
			width: data.length > 500 ? 1.5 : 2,
		},
	})

	let w        = chart.barWidth(),
		year     = (new Date(stats[0].date)).getFullYear(),
		lbl_btm  = window.label_bottom,
		lbl_top  = window.label_top,
		add      = (y, i) => {
				let s = document.createElement('span')
				s.style.left = `${i * w - 10}px`
				s.innerText = y
				lbl_btm.appendChild(s)
			},
		mark = (e, i) => {
				chart.draw(10, 0, 5, 100, function() {
					ctx.strokeStyle = 'rgba(0, 0, 0, .3)'
					ctx.lineWidth   = Math.max(chart.barWidth(), 3)

					let x = i*chart.barWidth() + chart.barWidth()/2
					ctx.beginPath()
					ctx.moveTo(x, 2.5)
					ctx.lineTo(x, 97)
					ctx.stroke()

					let s = document.createElement('span')
					s.style.left = `0px`
					s.title = e
					s.innerText = e.substr(e.indexOf(':') + 2)
					lbl_top.appendChild(s)

					let ww = Math.max(s.scrollWidth, s.offsetWidth, s.clientWidth)
					s.style.left = `${i*w - ww/2}px`

					if (ww > Math.max(s.scrollWidth, s.offsetWidth, s.clientWidth))
						s.style.left = `${i*w - ww}px`
				})
		}
	stats.forEach((s, i) => {
		/// Draw x-axis.
		let y = (new Date(s.date)).getFullYear()
		if (y !== year)
			add(y, i)
		year = y

		/// Mark forks.
		let e = events[s.date]
		if (e && e.slice(0, 4) !== 'tag:')
			mark(e, i)
	})

	// Show tooltip and highlight position on mouse hover.
	let tip    = document.createElement('div'),
		reset  = {x: -1, y: -1, f: () => {}},
		height = (e) => Math.max(e.scrollHeight, e.offsetHeight, e.clientHeight),
		width  = (e) => Math.max(e.scrollWidth, e.offsetWidth, e.clientWidth)
	tip.id = 'tooltip'
	chart.mouse(function(i, x, y, w, h, offset, ev) {
		if (ev == 'leave') {
			tip.remove()
			reset.f()
			return
		}
		else if (ev === 'enter') { }
		else if (x === reset.x)
			return

		let day   = stats[i],
			event = events[day.date]
		tip.remove()
		if (weekly)
			tip.innerHTML = `${format_date(day.date)} to ${format_date(day.date, 6)} - ${day.commits} commits, ${day.added} added, ${day.removed} removed`
		else
			tip.innerHTML = `${format_date(day.date)} - ${day.commits} commits, ${day.added} added, ${day.removed} removed`
		if (event)
			tip.innerHTML += ` – <b>${esc(event)}</b>`

		document.body.appendChild(tip)
		tip.style.left = `${offset.left + x}px`
		if (height(tip) > 40) {
			tip.style.left = '0px'
			tip.style.left = `${x + offset.left - width(tip) - 8}px`
		}
		tip.style.top  = `${offset.top - height(tip) - 10}px`

		reset.f()
		reset = chart.draw(x, 0, w, h, function() {
			ctx.strokeStyle = '#999'
			ctx.fillStyle   = 'rgba(99, 99, 99, .5)'
			ctx.lineWidth   = 1

			ctx.beginPath()
			ctx.moveTo(x + ctx.lineWidth/2, 2.5)
			ctx.lineTo(x + ctx.lineWidth/2, 140)
			ctx.stroke()
		})
	})
}

tsort()
draw_chart()
