function decodeBase64URL(value: string): ArrayBuffer {
  const base64 = value.replace(/-/g, '+').replace(/_/g, '/').padEnd(Math.ceil(value.length / 4) * 4, '=')
  const bytes = Uint8Array.from(atob(base64), char => char.charCodeAt(0))
  return bytes.buffer
}

function encodeBase64URL(value: ArrayBuffer): string {
  const bytes = new Uint8Array(value)
  let binary = ''
  bytes.forEach(byte => { binary += String.fromCharCode(byte) })
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export function creationOptions(options: any): CredentialCreationOptions {
  const publicKey = options.publicKey || options
  return { publicKey: {
    ...publicKey,
    challenge: decodeBase64URL(publicKey.challenge),
    user: { ...publicKey.user, id: decodeBase64URL(publicKey.user.id) },
    excludeCredentials: (publicKey.excludeCredentials || []).map((item:any) => ({ ...item, id: decodeBase64URL(item.id) })),
  } }
}

export function requestOptions(options: any): CredentialRequestOptions {
  const publicKey = options.publicKey || options
  return { publicKey: {
    ...publicKey,
    challenge: decodeBase64URL(publicKey.challenge),
    allowCredentials: (publicKey.allowCredentials || []).map((item:any) => ({ ...item, id: decodeBase64URL(item.id) })),
  } }
}

export function credentialJSON(credential: PublicKeyCredential): Record<string, unknown> {
  const response = credential.response
  const common = {
    id: credential.id,
    rawId: encodeBase64URL(credential.rawId),
    type: credential.type,
    clientExtensionResults: credential.getClientExtensionResults(),
  }
  if (response instanceof AuthenticatorAttestationResponse) {
    return { ...common, response: {
      clientDataJSON: encodeBase64URL(response.clientDataJSON),
      attestationObject: encodeBase64URL(response.attestationObject),
      transports: response.getTransports?.() || [],
    } }
  }
  const assertion = response as AuthenticatorAssertionResponse
  return { ...common, response: {
    clientDataJSON: encodeBase64URL(assertion.clientDataJSON),
    authenticatorData: encodeBase64URL(assertion.authenticatorData),
    signature: encodeBase64URL(assertion.signature),
    userHandle: assertion.userHandle ? encodeBase64URL(assertion.userHandle) : null,
  } }
}
