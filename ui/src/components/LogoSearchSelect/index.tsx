import React, { useState, useCallback, useRef } from "react";
import { AutoComplete, Spin } from "antd";

// Iconify 搜索 API
const ICONIFY_SEARCH_API = "https://api.iconify.design/search";

interface IconOption {
  value: string;      // 图标 SVG 的 URL
  label: React.ReactNode;
  iconName: string;
  svgUrl: string;
}

interface LogoSearchSelectProps {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
}

/**
 * 通过 Iconify 关键词搜索图标，选择后写入 logo 字段（存的是图标 SVG 的 URL）；
 * 同时允许手动输入任意 URL。
 */
const LogoSearchSelect: React.FC<LogoSearchSelectProps> = ({
  value,
  onChange,
  placeholder,
}) => {
  const [options, setOptions] = useState<IconOption[]>([]);
  const [open, setOpen] = useState(false);
  const [fetching, setFetching] = useState(false);
  const fetchIdRef = useRef(0);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // 远程搜索 iconify
  const doFetch = useCallback((keyword: string) => {
    const fetchId = ++fetchIdRef.current;
    setFetching(true);
    fetch(`${ICONIFY_SEARCH_API}?query=${encodeURIComponent(keyword)}&limit=20`)
      .then((res) => res.json())
      .then((data) => {
        if (fetchId !== fetchIdRef.current) return; // 丢弃过期请求
        const icons: string[] = data?.icons || [];
        const opts: IconOption[] = icons.slice(0, 20).map((icon) => {
          const svgUrl = `https://api.iconify.design/${icon}.svg`;
          return {
            value: svgUrl,
            iconName: icon,
            svgUrl,
            label: (
              <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                <img src={svgUrl} width={20} height={20} alt={icon} />
                <span style={{ fontSize: 12, color: "#666" }}>{icon}</span>
              </div>
            ),
          };
        });
        setOptions(opts);
      })
      .catch(() => {
        setOptions([]);
      })
      .finally(() => {
        if (fetchId === fetchIdRef.current) setFetching(false);
      });
  }, []);

  // 输入防抖后搜索
  const handleSearch = (keyword: string) => {
    if (timerRef.current) clearTimeout(timerRef.current);
    if (keyword && keyword.trim()) {
      setOpen(true);
      const kw = keyword.trim();
      timerRef.current = setTimeout(() => {
        doFetch(kw);
      }, 400);
    } else {
      setOptions([]);
    }
  };

  const handleSelect = (val: string) => {
    if (onChange) onChange(val);
    setOpen(false);
  };

  return (
    <AutoComplete
      value={value}
      onChange={onChange}
      onSearch={handleSearch}
      onSelect={handleSelect}
      onFocus={() => setOpen(true)}
      onBlur={() => setTimeout(() => setOpen(false), 200)}
      open={open}
      options={options.map((o) => ({ value: o.svgUrl, label: o.label }))}
      placeholder={placeholder || "输入关键词搜索图标，或直接输入 URL"}
      style={{ width: "100%" }}
      notFoundContent={
        fetching ? (
          <Spin size="small" />
        ) : (
          <div style={{ padding: 8, color: "#888" }}>无匹配图标，可直接输入 URL</div>
        )
      }
    />
  );
};

export default LogoSearchSelect;
