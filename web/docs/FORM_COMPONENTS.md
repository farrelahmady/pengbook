# Form Components — Panduan Komponen Form Terpadu

> Dibuat: 2026-09-17. Konteks: unifikasi 7 dropdown `<select>` + input + tombol
> submit yang sebelumnya dicopy-paste di form jurnal, akun, aset, dan demo.

## Latar Belakang

Setiap form mendefinisikan ulang string `selectClass` / `labelClass` / tombol
submit yang identik, plus blok grouping `Map<type, list>` dan render
`{code} · {name}` yang sama. Perubahan satu token desain harus diedit di
7 file — rawan divergen (terbukti: dropdown "To" di `basic-form` sempat flat
sementara "From" grouped).

## Arsitektur (3 Lapis)

```
AccountSelect (domain: akun CoA, grouping, "code · name", loading)
  └─ AppSelect (native <select>, 1 definisi style)
AppField (label + hint/error + wiring htmlFor/id)
AppInput (shadcn Input + token form project)
AppSubmitButton (varian sheet | form | stateful)
groupByType() di lib/group-accounts.ts (grouping display-only)
```

| File | Peran |
|------|-------|
| `components/ui/app-field.tsx` | Wrapper label + hint/error. Menggantikan `labelClass` copy-paste; memasangkan `label ↔ control` via `htmlFor/id` (aksesibilitas). |
| `components/ui/app-input.tsx` | shadcn `Input` dengan override token project (`rounded-xl bg-secondary-50 text-[13px]`). |
| `components/ui/app-select.tsx` | **Satu-satunya `<select>` di codebase.** Native, bukan Radix (lihat bawah). API `groups` 1:1 dengan Radix `SelectGroup`. |
| `components/ui/app-submit-button.tsx` | 3 varian submit: `sheet` (rounded-2xl + shadow), `form` (rounded-xl), `stateful` (biru/abu berdasar `ready`). |
| `components/akun/account-select.tsx` | Dropdown domain CoA. Terima `PostingAccount \| ParentListItem`, standarkan `value: string` (`Number()` hanya di boundary submit). |
| `lib/group-accounts.ts` | `groupByType()` — grouping untuk `<optgroup>`. Display-only, bukan kalkulasi bisnis (urutan & label dari payload API). |

Contoh pakai:

```tsx
<AccountSelect
  label={t("fromAccount")}
  value={fromAccount}
  onChange={setFrom}
  accounts={postingAccounts}
  isLoading={isLoadingAccounts}
  placeholder={t("fromPlaceholder")}
  loadingText="Loading..."
/>
```

## Mengapa AppSelect = Native, Bukan shadcn `Select` (Radix)?

Keputusan sadar, bukan keterbatasan. Empat masalah konkret jika memakai
`ui/select.tsx` bawaan saat itu:

1. **Regresi visual.** Token bawaan (`rounded-none text-xs h-8 w-fit
   bg-transparent`) bertolak belakang dengan tema form
   (`rounded-xl text-[13px] w-full py-3 bg-secondary-50`). Memakai mentah =
   dropdown kotak/kecil/menyusut; menyamakan = override total (melawan shadcn).
2. **Portal + Sheet.** `SelectContent` me-render via Portal (`z-50`) sementara
   `SheetContent` juga `fixed z-50` + focus trap (basis Dialog). Di bottom-sheet
   mobile: daftar bisa jatuh di bawah overlay, fokus rebutan (daftar langsung
   tertutup / Escape menutup Sheet), dan `position="item-aligned"` salah ukur
   saat animasi `slide-in-from-bottom` + terpotong `overflow-y-auto 92dvh`.
   Picker OS bawaan tidak punya masalah ini dan UX-nya lebih baik di mobile.
3. **Baris compact pecah.** Grid advanced-form `1fr_80px_80px_28px` vs trigger
   `w-fit` + konten `min-w-36` (144px) = overflow kolom 80px di layar 390px.
4. **Skala.** `Select` polos tanpa search tidak membantu saat akun ratusan
   (item `text-xs`, tanpa pencarian) — downgrade dari picker OS. Solusi skala
   yang benar adalah **Combobox (`Popover + Command`)**, komponen terpisah.

### Kapan shadcn `Select` tepat? (Fase 5, opsional)

Jika butuh render item kaya (ikon/badge), dipakai di dialog desktop terpusat
(bukan bottom-sheet), dan `SelectTrigger/Content/Item` sudah di-restlyle ke
token project. Prop `groups` di `AppSelect` sudah memetakan 1:1 ke
`SelectGroup + SelectLabel`, sehingga migrasi hanya mengganti isi
`app-select.tsx` tanpa menyentuh call-site. Jika butuh search saat akun
banyak, loncat ke Combobox, bukan `Select`.

### Restore primitif yang dihapus

Scaffold shadcn tak terpakai dihapus pada 2026-09-17 (lihat bawah).
Generate ulang bila dibutuhkan:

```bash
npx shadcn@latest add select accordion badge dialog popover separator switch tabs
```

## Pembersihan Komponen Tak Terpakai (2026-09-17)

Diverifikasi 0 import statis + 0 dynamic `import()` di seluruh `web/`.
Dihapus:

- `components/shared/money-text.tsx` (`MoneyText` tak dipakai siapa pun;
  hook `use-currency-formatter` tetap dipakai 8 file lain)
- `components/ui/accordion.tsx`, `badge.tsx`, `dialog.tsx`, `popover.tsx`,
  `select.tsx`, `separator.tsx`, `switch.tsx`, `tabs.tsx`

Tetap dipakai dan dipertahankan: `sheet`, `skeleton`, `sonner`,
`transfer-progress-toast`, `dropdown-menu`, `label` (via `AppField`),
`input` (via `AppInput`), `button` (via `sheet`/`dialog` internal).

## Aturan Kontribusi

- Butuh varian baru (misal input currency, textarea)? Tambah di primitif
  `app-*`, jangan copy-paste class ke form.
- Butuh dropdown domain baru? Ikuti pola `AccountSelect` (terima data mentah
  dari API, `value: string`, konversi tipe hanya di boundary submit).
- Grouping/filter untuk tampilan = boleh di frontend (presentation logic).
  Kalkulasi bisnis (harga, total, pajak) = tetap di `api/` (aturan utama).
