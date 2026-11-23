import { getLogoUrl } from './check';

/**
 * getLogoUrl 是 logo 显示的唯一入口，绕开它自己拼 URL 会踩两个坑：
 *   1. data: URI 进 /api/img → Go 1.17+ 把 ';' 当非法 query 分隔符，整段参数丢失 → 400 裂图
 *   2. 不编码 → base64 里的 '+' 被解码成空格 → 解码失败 → 裂图
 * 这些用例锁死这两条行为。
 */
describe('getLogoUrl', () => {
  const DATA_URI = 'data:image/svg+xml;base64,PHN2Zz48L3N2Zz4=';

  it('data: URI 原样返回，不过代理', () => {
    expect(getLogoUrl(DATA_URI)).toBe(DATA_URI);
  });

  it('含 + 和 / 的 base64 也原样返回（不做任何转义）', () => {
    const tricky = 'data:image/png;base64,AAA+BBB/CCC==';
    expect(getLogoUrl(tricky)).toBe(tricky);
  });

  it('http(s) 走 /api/img 代理并 encodeURIComponent', () => {
    const url = 'https://api.iconify.design/mdi/sim.svg';
    expect(getLogoUrl(url)).toBe(`/api/img?url=${encodeURIComponent(url)}`);
  });

  it('http URL 里的 + 被编码成 %2B（否则后端会解成空格）', () => {
    const url = 'https://a.com/x.svg?a=1+2';
    const out = getLogoUrl(url);
    expect(out).toContain('%2B');
    expect(out).not.toContain('1+2');
  });

  it('已是根相对路径的原样返回', () => {
    expect(getLogoUrl('/foo.png')).toBe('/foo.png');
  });

  it('普通文件名补前导斜杠', () => {
    expect(getLogoUrl('baidu.ico')).toBe('/baidu.ico');
  });

  it('空值返回空', () => {
    expect(getLogoUrl('')).toBe('');
  });
});
