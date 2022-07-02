import {amendNode, clearNode} from './lib/dom.js';
import {br, button, div, fieldset, legend, textarea} from './lib/html.js';
import {footer, header, registerPage} from './pages.js';
import {ready, rpc} from './rpc.js';
import {labels} from './shared.js';

const h = textarea(),
      f = textarea(),
      update = button({"onclick": () => {
	for (const e of elements) {
		amendNode(e, {"disabled": true});
	}
	const wp = window.open("", "", "");
	if (wp) {
		wp.addEventListener("unload", () => {
			for (const e of elements) {
				amendNode(e, {"disabled": false});
			}
		});
		const header = h.value,
		      footer = f.value;
		wp.document.documentElement.innerHTML = `${header}<div id="HEADERFOOTERTESTER"></div>${footer}`;
		const tester = wp.document.getElementById("HEADERFOOTERTESTER");
		if (tester) {
			tester.replaceWith(button({"onclick": () => {
				wp.close();
				rpc.setHeaderFooter(header, footer);
			}}, "Looks Good"), button({"onclick": () => wp.close()}, "Looks Bad"));
		} else {
			wp.close();
			alert("Invalid Header or Footer");
		}
	} else {
		alert("Need to allow pop-up windows");
		return;
	}
      }}, "Update"),
      clear = button({"onclick": () => {
		clearNode(h, header);
		clearNode(f, footer);
      }}, "Reset"),
      elements = [h, f, update, clear];

ready.then(() => {
	amendNode(h, header);
	amendNode(f, footer);
});

registerPage("headFoot", "Set Header/Footer", div(fieldset([
	legend("Set Header & Footer"),
	labels("Header: ", h),
	br(),
	labels("Footer: ", f),
	br(),
	update,
	clear
])));
