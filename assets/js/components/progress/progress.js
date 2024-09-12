import * as $ from "bm.js/bm.module.js";
import {html, render} from 'lit-html';
import {css} from "common.js";

/*
<c-progress value="0.5" height="0.3rem"><c-progress>
*/

class CustomElements extends $.CustomElement {

template() {
    return html`
        <style>
            ${css}

            :host {
                display: block;
                background-color: var(--gray-100);
                height: ${this.attr("height") || "0.3rem"};
            }
            #bar {
                width: ${this.attr("value") * 100}%;
                background-color: var(--green-400);
                height: ${this.attr("height") || "0.3rem"};
            }
        </style>
        <section>
            <section id="bar"></section>
        </section>
        `;
    }
	constructor() {
		super();
	}
	static get observedAttributes() {
		return [ "value" ];
	}
	
	async render() {
		render(this.template(), this.shadowRoot);
	}
}
customElements.define("c-progress", CustomElements);
