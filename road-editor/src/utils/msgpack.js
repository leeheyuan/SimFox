import { decode, encode } from '@msgpack/msgpack'

function base64ToUint8ArrayCross(base64){
  if (typeof window === 'undefined') {
    // Node.js
    return new Uint8Array(Buffer.from(base64, 'base64'));
  } else {
    // 浏览器
    const binary = atob(base64);
    return Uint8Array.from(binary, c => c.charCodeAt(0));
  }
}
 


function arrayBufferToUtf8String(buffer) {
  const decoder = new TextDecoder("utf-8");
  return decoder.decode(buffer);
}

export function parseMsgpack(base64) {    
  const buffer = base64ToUint8ArrayCross(arrayBufferToUtf8String(base64))  
  return decode(buffer)
}

export function parseMsgpackNoCode(buffer) {      
  return decode(buffer)
}

export function encodeMsgpack(data) {
  return encode(data)
}
 

/** 把 Blob 转换成 Uint8Array */
async function blobToUint8Array(blob) {
  const buffer = await blob.arrayBuffer(); // 现代浏览器支持
  return new Uint8Array(buffer);
}

/** 解码 Blob 类型的 msgpack 数据 */
export async function decodeMsgpackBlob(blob) {
  const uint8 = await blobToUint8Array(blob);
  return decode(uint8); // 这里就能正常解析成 JS 对象
}

