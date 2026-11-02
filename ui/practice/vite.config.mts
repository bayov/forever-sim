// The practice room's own Vite config (docs/practice_room.md). It builds ui/practice into
// dist/practice and shares nothing with the regular build's config, which writes dist/forever.
import path from 'path';
import { defineConfig } from 'vite';

const root = path.resolve(__dirname);

export default defineConfig({
	root,
	base: '/',
	publicDir: false,
	build: {
		outDir: path.resolve(root, '../../dist/practice'),
		// practice.wasm and wasm_exec.js live there too, from make practice-wasm.
		emptyOutDir: false,
		rollupOptions: { input: { practice: path.resolve(root, 'practice.html') } },
		target: 'es2022',
	},
	worker: { format: 'es' },
	server: {
		port: 3435,
		proxy: {
			'/api': 'http://127.0.0.1:3434',
			'/practice.wasm': 'http://127.0.0.1:3434',
			'/wasm_exec.js': 'http://127.0.0.1:3434',
		},
	},
});
