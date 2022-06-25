import {clearNode} from './lib/dom.js';
import {div, h1} from './lib/html.js';
import {body, ready} from './rpc.js';

ready.catch(e => {
	clearNode(body, [
		h1("Error"),
		div(e.message)
	]);
	throw e;
}).then(() => {
});
