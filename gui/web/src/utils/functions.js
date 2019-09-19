export function assetDisplay(code, issuer) {
    if (code === "XLM") {
        return code;
    }
    return code + ":" + issuer;
}

export function capSdexPrecision(num) {
    return num.toFixed(7);
}

export default { assetDisplay, capSdexPrecision };