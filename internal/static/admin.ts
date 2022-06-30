import {amendNode, clearNode} from './lib/dom.js';
import {div, h1, li, ul} from './lib/html.js';
import setHeaderFooter from './headfoot.js';
import {body, ready} from './rpc.js';

ready.catch(e => {
	clearNode(body, [
		h1("Error"),
		div(e.message ?? "Unknown Error")
	]);
	throw e;
}).then(() => {
	const contents = div();
	amendNode(body, [
		ul([
			li({"onclick": () => clearNode(contents, setHeaderFooter)}, "Set Header/Footer")
		]),
		contents
	]);
});
