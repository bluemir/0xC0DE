import * as $ from "bm.js/bm.module.js";

// WebAuthn 옵션과 응답의 base64url 변환은 브라우저 표준 메서드를 쓴다.
// PublicKeyCredential.parseCreationOptionsFromJSON / parseRequestOptionsFromJSON / toJSON
export function supported() {
	return !!(window.PublicKeyCredential
		&& PublicKeyCredential.parseCreationOptionsFromJSON
		&& PublicKeyCredential.parseRequestOptionsFromJSON);
}

function check() {
	if (!supported()) {
		throw new Error("이 브라우저는 passkey를 지원하지 않습니다");
	}
}

// register 는 passkey 를 만들어 서버에 등록한다.
// username 은 passkey 로 새로 가입할 때만 필요하다.
export async function register({username = "", label = ""} = {}) {
	check();

	let begin = await $.request("POST", "/api/v1/passkeys/register/begin", {
		body: {username},
	});

	let credential = await navigator.credentials.create(
		PublicKeyCredential.parseCreationOptionsFromJSON(begin.json.publicKey),
	);

	let res = await $.request("POST", "/api/v1/passkeys/register/finish", {
		query: {label},
		body: credential.toJSON(),
	});

	return res.json;
}

// login 은 username 없이 passkey 로 로그인한다.
export async function login() {
	check();

	let begin = await $.request("POST", "/api/v1/passkeys/login/begin");

	let credential = await navigator.credentials.get(
		PublicKeyCredential.parseRequestOptionsFromJSON(begin.json.publicKey),
	);

	let res = await $.request("POST", "/api/v1/passkeys/login/finish", {
		body: credential.toJSON(),
	});

	return res.json;
}

export async function list() {
	let res = await $.request("GET", "/api/v1/passkeys");
	return res.json.Items;
}

export async function revoke(index) {
	await $.request("DELETE", `/api/v1/passkeys/${index}`);
}
