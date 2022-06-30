import {amendNode, clearNode} from './lib/dom.js';
import {br, button, div, fieldset, legend, textarea} from './lib/html.js';
import {footer, header, ready} from './rpc.js';
import {labels} from './shared.js';

const h = textarea(),
      f = textarea();

ready.then(() => {
	amendNode(h, header);
	amendNode(f, footer);
});

export default div(fieldset([
	legend("Set Header & Footer"),
	labels("Header: ", h),
	br(),
	labels("Footer: ", f),
	br(),
	button({}, "Update"),
	button({"onclick": () => {
		clearNode(h, header);
		clearNode(f, footer);
	}}, "Reset")
]));
