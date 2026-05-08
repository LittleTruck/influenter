# PM Review — 優化項目清單

來源：杜宗祐 Oscar 於 2026-04-19 ~ 2026-05-08 的 review 回饋
整理日：2026-05-08
範圍：本文件僅收錄「需後續實作的優化／設計調整／新功能」。
Bug 修復另行於 branch `fix/pm-review-bugs` 直接處理，不在本清單。

---

## 分類索引

| 編號 | 類型 | 標題 | 優先級建議 |
| --- | --- | --- | --- |
| #3 | UI 連動 | 左下角頭像連動帳號頭像 | P2（小） |
| #4 | UI 樣式 | AI 助理編輯器接近 Gmail 排版 | P2 |
| #5b | 需求釐清 | 合作項目「描述」欄位用途 | P1（先釐清） |
| #6 | 欄位調整 | 專案頁面上方第一塊欄位 | P1 |
| #9 | 新功能 | 合作管理 → 數據分析頁 | P2 |
| #10 | UI 風格 | 整體較少框線的設計風格 | P3 |

> 備註：#1、#2、#5a（同步失敗）、#7、#8 屬於 bug，已在 `fix/pm-review-bugs` 直接修復。

---

## #3 — 左下角頭像連動帳號頭像

**現況**：左下角 sidebar 頭像未綁定使用者實際頭像。

**期望**：
- 顯示登入使用者的 Google 帳號頭像（OAuth profile picture）。
- 若無頭像則 fallback 為使用者名稱首字。

**實作要點**：
- 後端：確認 `users` 資料表是否儲存 `picture_url`（Google OAuth response 內的 `picture` 欄位）。若無，補上 migration 與 OAuth callback 寫入邏輯。
- 前端：`BaseDashboardSidebar.vue` 的左下角頭像 binding 改為從 `useAuth()` 取出當前 user 的 picture URL。
- 圖片 fallback 走 `BaseAvatar` 既有的首字邏輯。

**估時**：0.5 ~ 1 day

---

## #4 — AI 助理編輯器（三欄 + 回覆範本）排版接近 Gmail

**現況**：「那三欄」（AI 助理產出區塊）與「回覆範本」編輯器的行距、字體大小與 Gmail 不一致，閱讀體驗較差。

**期望**：
- 行距、字體大小、字距接近 Gmail compose / read pane 的觀感。
- 預設字體建議：`-apple-system, "Segoe UI", Roboto, sans-serif`，14px，`line-height: 1.5`。

**實作要點**：
- 找出對應的編輯器元件（推測為 `frontend/app/components/emails/` 下的 reply / template 相關元件 + AI 助理三欄區塊）。
- 用一個共用 class（如 `.gmail-like-editor`）統一樣式，避免散落。
- 若使用 ProseMirror / TipTap，於 editor extensions 設定 typography defaults。

**估時**：0.5 day

---

## #5b — 合作項目「描述」欄位的用途與位置

**問題**：PM 詢問「合作項目注意事項」這類資訊應寫在「描述」欄位，還是另開新欄位？

**狀態**：**待 PM 釐清**，再決定實作方向。

**建議討論方向**（給 PM 參考）：

| 方案 | 說明 | 優點 | 缺點 |
| --- | --- | --- | --- |
| A. 沿用「描述」欄 | 描述就是「給該合作項目套用時帶入的注意事項」 | 不增欄位 | 描述若同時要記內部備註會混淆 |
| B. 新增「合作項目注意事項」欄 | 描述＝面向品牌的對外說明；注意事項＝內部 KOL 自己提醒自己 | 語意清楚 | 多一個欄位 |
| C. 新增「Checklist」欄 | 結構化的待辦條列（例如必拍角度、必標 tag） | 可勾選追蹤 | 實作較重 |

**建議**：先採方案 B，欄位名「製作備註 / 注意事項」，型別為 rich text。

**估時**：定案後 0.5 day（DB migration + 前端欄位 + 顯示）

---

## #6 — 專案頁面上方第一塊可編輯項目調整

**測試案例**：`【Kolr商業合作】Uber Eats 應該都點得到_IG Reels 合作邀約 / 開根號`

**逐項調整**：

| 子項 | 現況 | 期望 |
| --- | --- | --- |
| 別名 | 顯示「別名」 | 改為「**代理商**」（label 改字，欄位本身可沿用既有 alias 欄） |
| 合作類型 | 獨立欄位 | 直接顯示已選的「報價項目」；若下方已有報價項目則此欄移除 |
| 總價 | 鎖定不可編輯 | 解除鎖定，可手動覆寫 |
| 聯絡人信箱 | 無 | **新增**此欄位 |

**實作要點**：
- 前端：`frontend/app/components/cases/detail/CasePropertiesPanel.vue`（推測）。
- 後端：若聯絡人信箱目前未持久化，需於 `cases` 資料表新增 `contact_email` 欄位 + migration。
  - 若已從 email 解析出 `from_email`，可考慮直接複用而非新欄位，與 PM 確認。
- 「總價解鎖」需確認原本鎖定原因是「自動由報價計算」嗎？若是，需在使用者手動編輯後加 `manually_overridden` flag，避免下次重新計算覆蓋。

**估時**：1 day（含 migration）

---

## #9 — 合作管理 → 新增「數據分析」頁

**參考**：CAPSULE Center 的圖表 + 表格樣式。

**規格**：
- **上方**：直條圖（堆疊長條）。
- **下方**：展開後可看詳細表格。
- **互動**：
  - X 軸固定為「月份」。
  - Y 軸由使用者切換：「金額」或「專案數」。
  - 堆疊維度固定為「合作項目」。

**實作要點**：
- 路由：`pages/analytics.vue` 或 `pages/cases/analytics.vue`。
- 後端：新增 `GET /api/v1/analytics/cases-by-month?metric=amount|count`，回傳 `{ month, items: [{ collaborationItem, value }] }`。
- 前端圖表：建議用既有 chart 套件（若無，採 Chart.js 或 ECharts；查詢 `frontend/package.json` 是否已有）。
- 表格部分用 `BaseTable` 直接渲染 raw 數據。

**估時**：2 ~ 3 day

---

## #10 — 整體 UI 風格：減少框線

**現況**：很多卡片、區塊有明顯 border，視覺較重。

**期望**：類似較少框線的設計（具體參考圖待 PM 補上）。

**實作方向**：
- 以 `box-shadow` 或淺背景色取代 border。
- 全域檢視 `BaseCard.vue`、`AppSection.vue`、`BasePanel.vue` 等，將預設 `border` 改為 `shadow-sm` 或無框線版本。
- 提供 `bordered` prop 保留向後相容（需要框線時仍可開啟）。

**注意**：本項牽涉廣，**先與設計師確認最終視覺**再動手，避免反覆改。

**估時**：1 ~ 2 day（視覺定稿後）

---

## 後續流程建議

1. **本週內**：先讓 PM 釐清 #5b（描述欄位定位）與 #10（具體參考圖）。
2. **下個 sprint**：依優先級先做 #3、#4、#6（皆為 P1 / P2 小調整）。
3. **後續 sprint**：#9 數據分析（中型功能）、#10 整體 UI 調整（與設計同步）。
