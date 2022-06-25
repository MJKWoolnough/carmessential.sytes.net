import {clearNode} from './lib/dom.js';
import {div, h1} from './lib/html.js';
import {ready} from './rpc.js';

ready.catch(e => {
	clearNode(document.body, [
		h1("Error"),
		div(e.message)
	]);
}).then(() => {
});
