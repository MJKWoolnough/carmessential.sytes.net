import {amendNode} from './lib/dom.js';
import {br, button, div, input, li, textarea, ul} from './lib/html.js';
import {NodeArray, NodeMap, node, stringSort} from './lib/nodes.js';
import {registerPage, setPage} from './pages.js';
import {labels} from './shared.js';
import {ready, rpc} from './rpc.js';

type Treatment = {
	name: string;
	price: number;
	description: string;
	duration: number;
}

type TreatmentNode = Treatment & {
	[node]: HTMLLIElement;
}

type Group = {
	group: string;
	arr: NodeArray<TreatmentNode, HTMLUListElement>;
	[node]: HTMLUListElement;
}

const treatmentSort = (a: Treatment, b: Treatment) => stringSort(a.name, b.name),
      contents = div();

ready.then(() => rpc.listTreatments().then(treatments => {
	const groups = new NodeMap<string, Group>(ul(), (a, b) => stringSort(a.group, b.group)),
	      treatmentName = input({"type": "text"}),
	      treatmentPrice = input({"type": "number", "step": "0.01", "min": 0}),
	      treatmentDescription = textarea(),
	      treatmentDuration = input({"type": "number", "step": 1, "min": 1, "value": 1}),
	      submitTreatment = button({"onclick": function(this: HTMLButtonElement) {
	      }}, "Create Treatment"),
	      noTreatment = {
		"name": "",
		"price": 0,
		"description": "",
		"duration": 1
	      },
	      setTreatment = (treatment: Treatment = noTreatment) => {
		      amendNode(treatmentName, {"value": treatment.name});
		      amendNode(treatmentPrice, {"value": treatment.price / 100});
		      amendNode(treatmentDescription, {"value": treatment.description});
		      amendNode(treatmentDuration, {"value": treatment.duration});
		      currTreatment = treatment;
		      setPage("setTreatment");
	      };
	let currTreatment: Treatment = noTreatment;
	for (const [_id, name, group, price, description, duration]  of treatments) {
		if (!groups.has(group)) {
			const arr = new NodeArray<TreatmentNode, HTMLUListElement>(ul(), treatmentSort);
			groups.set(group, {
				arr,
				group,
				[node]: arr[node]
			});
		}
		groups.get(group)?.arr.push({
			name, 
			price,
			description,
			duration,
			[node]: li(name)
		});
	}
	amendNode(contents, [
		button({"onclick": () => setTreatment()}, "New Treatment"),
		groups[node]
	]);
	registerPage("setTreatment", "", div([
		labels("Treatment Name: ", treatmentName),
		br(),
		labels("Treatment Price (£): ", treatmentPrice),
		br(),
		labels("Treatment Description: ", treatmentDescription),
		br(),
		labels("Treatment Duration (m): ", treatmentDuration),
		br(),
		submitTreatment
	]), () => {
		if (currTreatment.name !== treatmentName.value || currTreatment.price !== parseFloat(treatmentPrice.value) * 100 || currTreatment.description !== treatmentDescription.value || currTreatment.duration !== parseInt(treatmentDuration.value)) {
			if (!confirm("There are unsaved changes, are you sure you wish to change page?")) {
				return Promise.reject();
			}
		}
		return Promise.resolve();
	});
}));

registerPage("treatments", "Edit Treatments", contents);
