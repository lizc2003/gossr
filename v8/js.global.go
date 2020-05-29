package v8

const globalJsContent = `
this["console"] = {
	debug(...args) {
		v8worker.print(0, ...args)
	},
	log(...args) {
		v8worker.print(1, ...args)
	},
	info(...args) {
		v8worker.print(1, ...args)
	},
	warn(...args) {
		v8worker.print(2, ...args)
	},
	error(...args) {
		v8worker.print(3, ...args)
	}
};
`
