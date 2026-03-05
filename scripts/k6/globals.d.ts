declare module "https://jslib.k6.io/k6-utils/1.4.0/index.js" {
  export function uuidv4(): string;
}

declare module "https://jslib.k6.io/url/1.0.0/index.js" {
  export class URL {
    constructor(url: string);
  }
}
