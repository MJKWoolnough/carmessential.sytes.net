import parseBBCode from './lib/bbcode.js';
import {all} from './lib/bbcode_tags.js';
import {amendNode, clearNode} from './lib/dom.js';
import {br, button, datalist, div, h1, input, li, option, span, textarea, ul} from './lib/html.js';
import {NodeMap, node, stringSort} from './lib/nodes.js';
import {footer, header, registerPage, setPage} from './pages.js';
import {ready, rpc} from './rpc.js';
import {labels} from './shared.js';

const contents = div();

ready.then(() => rpc.listTreatments().then(treatments => {
	class Treatment {
		id: number;
		#name: string;
		#nameSpan: HTMLSpanElement;
		#group: string;
		price: number;
		description: string;
		duration: number;
		[node]: HTMLLIElement;
		constructor(id = -1, name = "", group = "", price = 0, description = "", duration = 1) {
			this.id = id;
			this[node] = li([
				this.#nameSpan = span(this.#name = name),
				button({"onclick": () => {}}, "Book"),
				button({"onclick": () => setTreatment(this)}, "Edit"),
				button({"onclick": () => {
					if (confirm("Are you sure you wish to remove this treatment?")) {
						rpc.removeTreatment(this.id).then(() => removeTreatmentFromGroup(this));
					}
				}}, "Remove")
			]);
			this.#group = group;
			this.price = price;
			this.description = description;
			this.duration = duration;
		}
		get name() {
			return this.#name;
		}
		set name(n: string) {
			clearNode(this.#nameSpan, this.#name = n);
		}
		get group() {
			return this.#group;
		}
		set group(g: string) {
			if (this.#group !== g) {
				removeTreatmentFromGroup(this);
				getGroup(this.#group = g).mp.set(this.id, this);
			}
		}
	}

	type Group = {
		group: string;
		mp: NodeMap<number, Treatment, HTMLUListElement>;
		[node]: HTMLLIElement;
	}

	const treatmentSort = (a: Treatment, b: Treatment) => stringSort(a.name, b.name),
	      groups = new NodeMap<string, Group>(ul(), (a, b) => stringSort(a.group, b.group)),
	      groupList = new NodeMap<string, {"id": string; [node]: HTMLOptionElement}, HTMLDataListElement>(datalist({"id": "groupNames"})),
	      treatmentTitle = h1(),
	      treatmentName = input({"type": "text"}),
	      treatmentGroup = input({"type": "text", "list": "groupNames"}),
	      treatmentPrice = input({"type": "number", "step": "0.01", "min": 0}),
	      treatmentDescription = textarea(),
	      treatmentDuration = input({"type": "number", "step": 1, "min": 1, "value": 1}),
	      submitTreatment = button({"onclick": function(this: HTMLButtonElement) {
		if (!treatmentName.value) {
			alert("Need a Name");
		} else if (!treatmentGroup.value) {
			alert("Need a Group");
		} else if (!treatmentPrice.value) {
			alert("Need a Price");
		} else if (!treatmentDescription.value) {
			alert("Need a Description");
		} else if (!treatmentDuration.value) {
			alert("Need a Duration");
		} else {
			const price = Math.floor(parseFloat(treatmentPrice.value) * 100),
			      duration = parseInt(treatmentDuration.value);
			amendNode(submitTreatment, {"disabled": true});
			(currTreatment.id === -1 ? rpc.addTreatment(treatmentName.value, treatmentGroup.value, price, treatmentDescription.value, duration).then(id => addTreatment(currTreatment = new Treatment(id, treatmentName.value, treatmentGroup.value, price, treatmentDescription.value, duration))) : rpc.setTreatment(currTreatment.id, treatmentName.value, treatmentGroup.value, price, treatmentDescription.value, duration).then(() => {
				currTreatment.name = treatmentName.value;
				currTreatment.group = treatmentGroup.value;
				currTreatment.price = price;
				currTreatment.description = treatmentDescription.value;
				currTreatment.duration = duration;
			}))
			.then(() => setPage("treatments"))
			.catch(err => alert("Error: " + err))
			.finally(() => amendNode(submitTreatment, {"disabled": false}));
		}
	      }}),
	      setTreatment = (treatment: Treatment) => {
		treatmentName.value = treatment.name;
		treatmentGroup.value = treatment.group;
		treatmentPrice.value = (treatment.price / 100) + "";
		treatmentDescription.value = treatment.description;
		treatmentDuration.value = treatment.duration + "";
		currTreatment = treatment;
		const title = treatment.id === -1 ? "Create Treatment" : "Edit Treatment";
		clearNode(treatmentTitle, title);
		clearNode(submitTreatment, title);
		setPage("setTreatment");
	      },
	      getGroup = (group: string) => {
		let g = groups.get(group);
		if (!g) {
			const mp = new NodeMap<number, Treatment, HTMLUListElement>(ul(), treatmentSort);
			groupList.set(group, {"id": group, [node]: option({"value": group})});
			groups.set(group, g = {
				mp,
				group,
				[node]: li([
					group,
					mp[node]
				])
			});
		}
		return g;
	      },
	      removeTreatmentFromGroup = (t: Treatment) => {
			const group = groups.get(t.group);
			if (group) {
				group.mp.delete(t.id);
				if (!group.mp.size) {
					groups.delete(t.group);
					groupList.delete(t.group);
				}
			}
	      },
	      addTreatment = (treatment: Treatment) => getGroup(treatment.group).mp.set(treatment.id, treatment);
	let currTreatment: Treatment;
	for (const {id, name, group, price, description, duration}  of treatments) {
		addTreatment(new Treatment(id, name, group, price, description, duration));
	}
	amendNode(contents, [
		button({"onclick": () => setTreatment(new Treatment())}, "New Treatment"),
		groups[node]
	]);
	registerPage("setTreatment", "", div([
		treatmentTitle,
		labels("Treatment Name: ", treatmentName),
		br(),
		groupList[node],
		labels("Treatment Group: ", treatmentGroup),
		br(),
		labels("Treatment Price (£): ", treatmentPrice),
		br(),
		labels("Treatment Description: ", treatmentDescription),
		button({"onclick": () => {
			const wp = window.open("", "", "");
			if (wp) {
				wp.document.documentElement.innerHTML = `${header}<div id="TREATMENTTESTER"></div>${footer}`;
				const tester = wp.document.getElementById("TREATMENTTESTER");
				if (tester) {
					tester.replaceWith(parseBBCode(all, treatmentDescription.value));
				} else {
					wp.close();
					alert("Invalid Header or Footer");
				}
			} else {
				alert("Preview popup failed");
			}
		}}, "Preview"),
		br(),
		labels("Treatment Duration (m): ", treatmentDuration),
		br(),
		submitTreatment
	]), () => {
		if (currTreatment.name !== treatmentName.value || currTreatment.group !== treatmentGroup.value || currTreatment.price !== parseFloat(treatmentPrice.value) * 100 || currTreatment.description !== treatmentDescription.value || currTreatment.duration !== parseInt(treatmentDuration.value)) {
			if (!confirm("There are unsaved changes, are you sure you wish to change page?")) {
				return false;
			}
		}
		return true;
	});
}));

registerPage("treatments", "Edit Treatments", contents);
