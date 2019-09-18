export function assetDisplay(code, issuer) {
    if (code === "XLM") {
        return code;
    }
    return code + ":" + issuer;
}

export default { assetDisplay };