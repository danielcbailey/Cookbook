

export function setToken(token: string) {
    document.cookie = 'session='+token+'; path=/;'
}

export function invalidateToken() {
    document.cookie = 'session=; max-age=0; path=/;'
}

export function tokenValid(): boolean {
    const name = "session";
    const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    return !!match;
}