import {amendNode} from './lib/dom.js';
import {div, li, ul} from './lib/html.js';
import {NodeArray, NodeMap, node, stringSort} from './lib/nodes.js';
import {registerPage} from './pages.js';
import {ready, rpc} from './rpc.js';

type Treatment = {
	name: string;
	[node]: HTMLLIElement;
}

type Group = {
	group: string;
	arr: NodeArray<Treatment, HTMLUListElement>;
	[node]: HTMLUListElement;
}

const treatmentSort = (a: Treatment, b: Treatment) => stringSort(a.name, b.name),
      contents = div();

ready.then(() => rpc.listTreatments().then(treatments => {
	const groups = new NodeMap<string, Group>(ul(), (a, b) => stringSort(a.group, b.group))
	for (const [_id, name, group, _price, _description, _duration]  of treatments) {
		if (!groups.has(group)) {
			const arr = new NodeArray<Treatment, HTMLUListElement>(ul(), treatmentSort);
			groups.set(group, {
				arr,
				group,
				[node]: arr[node]
			});
		}
		groups.get(group)?.arr.push({
			name, 
			[node]: li(name)
		});
	}
	amendNode(contents, groups[node]);
}));

registerPage("treatments", "Edit Treatments", contents);
