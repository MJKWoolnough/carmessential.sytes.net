import type {Children, Props} from './lib/dom.js';
import {amendNode} from './lib/dom.js';
import {label} from './lib/html.js';

export const labels = (() => {
	type Input = HTMLInputElement | HTMLButtonElement | HTMLTextAreaElement | HTMLSelectElement;

	type LProps = Exclude<Props, NamedNodeMap>;

	interface Labeller {
		<T extends Input>(name: Children, input: T, props?: LProps): [HTMLLabelElement, T];
		<T extends Input>(input: T, name: Children, props?: LProps): [T, HTMLLabelElement];
	}

	let next = 0;
	return ((name: Children | Input, input: Input | Children, props: LProps = {}) => {
		const iProps = {"id": props["for"] = `ID_${next++}`};
		return name instanceof HTMLInputElement || name instanceof HTMLButtonElement || name instanceof HTMLTextAreaElement || name instanceof HTMLSelectElement ? [amendNode(name, iProps), label(props, input)] : [label(props, name), amendNode(input as Input, iProps)];
	}) as Labeller;
})();
