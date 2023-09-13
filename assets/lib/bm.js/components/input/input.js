import * as $ from "../../bm.module.js";
import {html, render} from 'lit-html';
import {css} from "../common.js";

var tmpl = (elem) => html`
	<style>
		${css}

		:host {
			position: relative;
			padding-top: 0.5rem;
		}
		input {
			display: block;
			outline: none;
			height: 2rem;
			border: 0;
			border-bottom: 1px solid var(--gray-200);
		}
		input:focus {
			border-bottom: 1px solid var(--gray-800);
		}

		input::placeholder {
			opacity: 0;
		}
		input:focus::placeholder {
			opacity: 1;
		}
		label {
			display: block;
			position: absolute;
			top: 0.5rem;
		}
		input:focus + label, input:not(:placeholder-shown) + label {
			font-size: 0.5rem;
			top: 0;
		}
	</style>
	<input id="input" type="${elem.attr("type")}" name="${elem.name}" placeholder="${elem.attr("placeholder") || " " }" @input="${ evt => elem._internal.setFormValue(evt.target.value)}" />
	<label for="input">${elem.attr("label")}</label>
`;

// TODO
// - animation
// - proxy attribute of input
// - type of input(outline, underline, ...)

class Input extends $.CustomElement {
	static get formAssociated() {
		return true;
	}

	static get observedAttributes() {
		return ["name", "label", "placeholder", "type"];
	}

	constructor() {
		super();

		this._internal = this.attachInternals();
	}
	async render() {
		render(tmpl(this), this.shadow);
	}

	get name(){
		return this.attr("name");
	}
}
customElements.define("c-input", Input);
