export function downloadAttachment(blob: Blob, contentDisposition: string | undefined, fallbackName: string) {
    const encodedName = contentDisposition?.match(/filename\*=utf-8''([^;]+)/i)?.[1]
    const plainName = contentDisposition?.match(/filename="?([^";]+)"?/i)?.[1]
    let fileName = fallbackName
    try {
        if (encodedName) fileName = decodeURIComponent(encodedName.trim())
        else if (plainName) fileName = decodeURIComponent(plainName.trim())
    } catch {
        fileName = fallbackName
    }
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = fileName
    anchor.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 0)
}