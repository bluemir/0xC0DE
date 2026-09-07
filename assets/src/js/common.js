import * as $ from "bm.js/bm.module.js";

let rev = $.get(`head script[type="importmap"]`).attr("rev");

export let css = `
@import url("/static/${rev}/css/element.css");
@import "/static/${rev}/bundle/fonts/fonts.css";
`;


export function closeDialog(evt) {
	if (evt.target.nodeName === 'DIALOG') {
		evt.target.close();
	}
}

// errorMessage 는 Error 와 $.request 의 실패 응답에서 사람이 읽을 메세지를 뽑는다.
export function errorMessage(e) {
	if (e instanceof Error) {
		return e.message;
	}

	// 서버는 RFC 9457 problem+json 으로 응답한다.
	try {
		let problem = JSON.parse(e.text);
		return problem.detail || problem.title;
	} catch {
		return e.text || `error: ${e.statusCode}`;
	}
}
