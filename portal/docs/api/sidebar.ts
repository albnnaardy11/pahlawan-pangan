import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/pahlawan-pangan-public-api",
    },
    {
      type: "category",
      label: "UNTAGGED",
      items: [
        {
          type: "doc",
          id: "api/cari-makanan-murah-b-2-c",
          label: "Cari Makanan Murah (B2C)",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/posting-makanan-surplus-provider-only",
          label: "Posting Makanan Surplus (Provider Only)",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/request-kurir-penyelamat",
          label: "Request Kurir Penyelamat",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/laporan-dampak-karbon-b-2-b",
          label: "Laporan Dampak Karbon (B2B)",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/cari-titik-drop-bersama-rt-rw",
          label: "Cari Titik Drop Bersama (RT/RW)",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/laporan-analitik-bisnis-roi",
          label: "Laporan Analitik Bisnis (ROI)",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
