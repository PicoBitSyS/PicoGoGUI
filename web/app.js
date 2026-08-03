(() => {
  const app = document.getElementById("app");
  const nodes = new Map();
  let modalRoot = null;
  let activeDialog = null;

  function post(msg) {
    const data = JSON.stringify(msg);
    if (typeof window.picoGoSend === "function") {
      window.picoGoSend(data);
      return true;
    }
    return false;
  }

  function emit(id, name, value) {
    const payload = { target: id, name };
    if (typeof value !== "undefined") payload.value = value;
    post({ kind: "event", id, event: name, payload });
  }

  const rpcHandlers = Object.create(null);
  window.__picoRegisterRPC = function (method, handler) {
    if (typeof method !== "string" || typeof handler !== "function") return;
    rpcHandlers[method] = handler;
  };

  async function handleRPC(msg) {
    const handler = rpcHandlers[msg.event];
    if (!handler) {
      post({ kind: "error", id: msg.id, payload: { message: "RPC method not found: " + (msg.event || "") } });
      return;
    }
    try {
      const value = await handler(msg.payload);
      post({ kind: "response", id: msg.id, payload: value });
    } catch (error) {
      post({ kind: "error", id: msg.id, payload: {
        message: error && error.message ? error.message : String(error)
      } });
    }
  }

  window.__picoRegisterRPC("runtime.info", () => ({
    userAgent: navigator.userAgent,
    language: navigator.language,
    theme: document.documentElement.getAttribute("data-theme") || "light"
  }));

  const customThemeVariables = [
    "--pico-bg", "--pico-fg", "--pico-muted", "--pico-accent",
    "--pico-control-bg", "--pico-control-hover", "--pico-input-bg",
    "--pico-border", "--pico-border-bottom", "--pico-radius",
    "--pico-gap", "--pico-pad", "--pico-control-width"
  ];

  function setTheme(spec) {
    const name = typeof spec === "string" ? spec : spec && spec.name;
    const variables = spec && typeof spec === "object" ? spec.variables : null;
    document.documentElement.setAttribute("data-theme", name || "light");
    for (const key of customThemeVariables) {
      document.documentElement.style.removeProperty(key);
    }
    if (variables) {
      for (const [key, value] of Object.entries(variables)) {
        if (customThemeVariables.includes(key) && value) {
          document.documentElement.style.setProperty(key, value);
        }
      }
    }
  }

  function clear(el) {
    while (el.firstChild) el.removeChild(el.firstChild);
  }

  function ensureModalRoot() {
    if (modalRoot) return modalRoot;
    modalRoot = document.createElement("div");
    modalRoot.id = "pico-modal-root";
    document.body.appendChild(modalRoot);
    return modalRoot;
  }

  function closeDialog() {
    const root = ensureModalRoot();
    if (activeDialog && activeDialog.keyHandler) {
      document.removeEventListener("keydown", activeDialog.keyHandler);
    }
    const previousFocus = activeDialog && activeDialog.previousFocus;
    activeDialog = null;
    clear(root);
    root.classList.remove("pico-modal-open");
    if (previousFocus && typeof previousFocus.focus === "function") previousFocus.focus();
  }

  function openDialog(spec) {
    if (!spec || !spec.id) return;
    const root = ensureModalRoot();
    clear(root);
    root.classList.add("pico-modal-open");

    const backdrop = document.createElement("div");
    backdrop.className = "pico-modal-backdrop";

    const panel = document.createElement("div");
    panel.className = "pico-modal";
    panel.setAttribute("role", "dialog");
    panel.setAttribute("aria-modal", "true");

    const title = document.createElement("div");
    title.className = "pico-modal-title";
    title.textContent = spec.title || "";

    const body = document.createElement("div");
    body.className = "pico-modal-body";
    body.textContent = spec.body || "";

    const actions = document.createElement("div");
    actions.className = "pico-modal-actions";
    const buttons = Array.isArray(spec.buttons) ? spec.buttons : ["OK"];

    let input = null;
    if (spec.kind === "prompt") {
      input = document.createElement("input");
      input.type = "text";
      input.className = "pico-textbox pico-modal-input";
      input.value = spec.value || "";
      if (spec.placeholder) input.placeholder = spec.placeholder;
      input.setAttribute("aria-label", spec.inputLabel || spec.title || "Value");
    }

    function finish(button) {
      const ok = button === "OK";
      emit(spec.id, "dialog", { ok, button, value: input ? input.value : undefined });
      closeDialog();
    }

    for (const label of buttons) {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = label === "OK" ? "pico-button pico-button-accent" : "pico-button";
      btn.textContent = label;
      btn.addEventListener("click", () => finish(label));
      actions.appendChild(btn);
    }

    panel.appendChild(title);
    panel.appendChild(body);
    if (input) panel.appendChild(input);
    panel.appendChild(actions);
    backdrop.appendChild(panel);
    root.appendChild(backdrop);
    backdrop.addEventListener("click", (ev) => {
      if (ev.target === backdrop && spec.dismissible !== false) finish("Cancel");
    });
    const keyHandler = (ev) => {
      if (ev.key === "Escape" && spec.dismissible !== false) {
        ev.preventDefault();
        finish("Cancel");
        return;
      }
      if (ev.key !== "Tab") return;
      const focusable = Array.from(panel.querySelectorAll("button, input, select, textarea, [tabindex]:not([tabindex='-1'])"))
        .filter((item) => !item.disabled);
      if (!focusable.length) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (ev.shiftKey && document.activeElement === first) {
        ev.preventDefault();
        last.focus();
      } else if (!ev.shiftKey && document.activeElement === last) {
        ev.preventDefault();
        first.focus();
      }
    };
    activeDialog = { keyHandler, previousFocus: document.activeElement };
    document.addEventListener("keydown", keyHandler);
    const focusTarget = input || actions.querySelector("button.pico-button-accent") || actions.querySelector("button");
    if (focusTarget) focusTarget.focus();
  }

  function renderTable(el, props, id) {
    props = Object.assign({}, el._picoTableProps || {}, props || {});
    el._picoTableProps = props;
    clear(el);
    const columns = Array.isArray(props.columns) ? props.columns : [];
    const rows = Array.isArray(props.rows) ? props.rows : [];
    const selected = props.selected;
    const filter = String(props.filter || "").trim().toLocaleLowerCase();
    const sortable = props.sortable === true;
    const state = el._picoTableSort || { column: "", ascending: true };
    let visibleRows = rows.map((row, index) => ({ row, index }));
    if (filter) {
      visibleRows = visibleRows.filter(({ row }) =>
        columns.some((col) => String(row && row[col] != null ? row[col] : "")
          .toLocaleLowerCase().includes(filter)));
    }
    if (state.column) {
      visibleRows.sort((a, b) => {
        const av = a.row && a.row[state.column];
        const bv = b.row && b.row[state.column];
        const result = String(av == null ? "" : av).localeCompare(
          String(bv == null ? "" : bv), undefined, { numeric: true, sensitivity: "base" });
        return state.ascending ? result : -result;
      });
    }

    const table = document.createElement("table");
    table.className = "pico-table-el";
    table.setAttribute("role", "grid");
    const thead = document.createElement("thead");
    const hr = document.createElement("tr");
    for (const col of columns) {
      const th = document.createElement("th");
      th.setAttribute("scope", "col");
      if (sortable) {
        const sort = document.createElement("button");
        sort.type = "button";
        sort.className = "pico-table-sort";
        const active = state.column === col;
        sort.textContent = col + (active ? (state.ascending ? " \u2191" : " \u2193") : "");
        sort.setAttribute("aria-sort", active ? (state.ascending ? "ascending" : "descending") : "none");
        sort.addEventListener("click", () => {
          const ascending = state.column === col ? !state.ascending : true;
          el._picoTableSort = { column: col, ascending };
          emit(id, "sort", { column: col, ascending });
          renderTable(el, props, id);
        });
        th.appendChild(sort);
      } else {
        th.textContent = col;
      }
      hr.appendChild(th);
    }
    thead.appendChild(hr);
    table.appendChild(thead);

    const tbody = document.createElement("tbody");
    visibleRows.forEach(({ row, index }) => {
      const tr = document.createElement("tr");
      if (selected === index) tr.classList.add("pico-selected");
      tr.tabIndex = 0;
      tr.setAttribute("aria-selected", selected === index ? "true" : "false");
      tr.addEventListener("click", () => emit(id, "select", index));
      tr.addEventListener("dblclick", () => emit(id, "activate", index));
      tr.addEventListener("keydown", (ev) => {
        if (ev.key === "Enter" || ev.key === " ") {
          ev.preventDefault();
          emit(id, "select", index);
        }
      });
      for (const col of columns) {
        const td = document.createElement("td");
        const cell = row && Object.prototype.hasOwnProperty.call(row, col) ? row[col] : "";
        td.textContent = cell == null ? "" : String(cell);
        tr.appendChild(td);
      }
      tbody.appendChild(tr);
    });
    table.appendChild(tbody);
    el.appendChild(table);
  }

  function snap(n, grid) {
    return Math.round(n / grid) * grid;
  }

  function num(v, fallback) {
    const n = Number(v);
    return Number.isFinite(n) ? n : fallback;
  }

  function isContainerKind(kind) {
    return kind === "column" || kind === "row" || kind === "stack";
  }

  function defaultWidgetSize(kind) {
    switch (kind) {
      case "label": return { w: 80, h: 22 };
      case "button": return { w: 100, h: 32 };
      case "textbox":
      case "numberbox":
      case "combobox": return { w: 160, h: 28 };
      case "checkbox": return { w: 120, h: 24 };
      case "badge": return { w: 72, h: 22 };
      case "column":
      case "row":
      case "stack": return { w: 220, h: 160 };
      default: return { w: 100, h: 32 };
    }
  }

  function applyAppearance(el, appearance) {
    if (!el) return;
    const a = appearance || {};
    el.style.fontFamily = a.fontFamily || "";
    el.style.fontSize = Number(a.fontSize) > 0 ? Number(a.fontSize) + "px" : "";
    el.style.color = a.color || "";
    el.style.backgroundColor = a.background || "";
    el.style.fontWeight = a.bold ? "700" : "";
    el.style.fontStyle = a.italic ? "italic" : "";
    el.style.textDecoration = a.underline ? "underline" : "";
    el.style.textAlign = a.textAlign || "";
    el.style.justifyContent =
      a.textAlign === "center" ? "center" :
      a.textAlign === "right" ? "flex-end" :
      a.textAlign === "left" ? "flex-start" : "";
    el.style.borderColor = a.borderColor || "";
    el.style.borderWidth = Number(a.borderWidth) > 0 ? Number(a.borderWidth) + "px" : "";
    el.style.borderStyle = Number(a.borderWidth) > 0 ? "solid" : "";
    el.style.borderRadius = Number(a.borderRadius) > 0 ? Number(a.borderRadius) + "px" : "";
    el.style.opacity = Number(a.opacity) > 0 ? String(Math.min(1, Number(a.opacity))) : "";
  }

  function fillWidgetContent(item, w, kind) {
    if (isContainerKind(kind)) {
      const body = document.createElement("div");
      body.className = "pico-ds-container-body";
      const caption = document.createElement("div");
      caption.className = "pico-ds-container-caption";
      caption.textContent = (w && w.text) || kind;
      item.appendChild(caption);
      item.appendChild(body);
      return body;
    }
    const pluginControls = window.__picoPlugins && window.__picoPlugins.controls;
    const pluginSpec = pluginControls && pluginControls[kind];
    if (pluginSpec && typeof pluginSpec.create === "function") {
      const props = Object.assign({}, w || {}, {
        text: (w && w.text) || kind,
        value: w && w.value,
        tone: (w && w.value) || "info"
      });
      const preview = pluginSpec.create({ id: (w && w.id) || "", kind, props }, props, { emit: () => {} });
      if (preview) {
        preview.style.pointerEvents = "none";
        item.appendChild(preview);
        return null;
      }
    }
    if (kind === "badge") {
      const tone = (w && w.value) || "info";
      item.classList.add("pico-ds-badge-" + tone);
      item.appendChild(document.createTextNode((w && w.text) || "Badge"));
      return null;
    }
    if (kind === "checkbox") {
      const box = document.createElement("span");
      box.className = "pico-ds-check";
      const lab = document.createElement("span");
      lab.textContent = (w && w.text) || "CheckBox";
      item.appendChild(box);
      item.appendChild(lab);
      return null;
    }
    if (kind === "textbox" || kind === "numberbox") {
      item.appendChild(document.createTextNode((w && w.value) || ""));
      return null;
    }
    if (kind === "combobox") {
      const txt = document.createElement("span");
      const parts = String((w && w.value) || "").split(",").map((s) => s.trim()).filter(Boolean);
      txt.textContent = parts[0] || (w && w.text) || "ComboBox";
      const arrow = document.createElement("span");
      arrow.className = "pico-ds-combo-arrow";
      arrow.textContent = "▾";
      item.appendChild(txt);
      item.appendChild(arrow);
      return null;
    }
    if (kind === "button") {
      item.appendChild(document.createTextNode((w && w.text) || "Button"));
      return null;
    }
    item.appendChild(document.createTextNode((w && w.text) || "Label"));
    return null;
  }

  function addResizeHandles(item) {
    ["nw", "ne", "sw", "se"].forEach((dir) => {
      const h = document.createElement("span");
      h.className = "pico-ds-handle pico-ds-handle-" + dir;
      h.dataset.handle = dir;
      item.appendChild(h);
    });
  }

  function designSelection(props) {
    const raw = Array.isArray(props && props.selection) ? props.selection : [];
    const out = [];
    for (const value of raw) {
      const index = Number(value);
      if (Number.isInteger(index) && index >= 0 && !out.includes(index)) out.push(index);
    }
    if (out.length === 0 && Number.isInteger(props && props.selected) && props.selected >= 0) {
      out.push(props.selected);
    }
    return out;
  }

  function directDesignWidgets(parentEl) {
    return Array.from((parentEl && parentEl.children) || []).filter((node) =>
      node.classList && node.classList.contains("pico-ds-widget")
    );
  }

  function updateDesignSelection(hostEl, selection) {
    const primary = selection.length ? selection[selection.length - 1] : -1;
    hostEl.querySelectorAll(".pico-ds-widget").forEach((node) => {
      const index = Number(node.dataset.index);
      node.classList.toggle("pico-ds-selected", selection.includes(index));
      node.classList.toggle("pico-ds-primary", index === primary);
      node.querySelectorAll(".pico-ds-handle").forEach((handle) => handle.remove());
      if (index === primary && node.dataset.locked !== "true") addResizeHandles(node);
    });
  }

  function clearDesignGuides(parentEl) {
    if (!parentEl) return;
    Array.from(parentEl.children || []).forEach((node) => {
      if (node.classList && node.classList.contains("pico-ds-guide")) node.remove();
    });
  }

  function addDesignGuide(parentEl, vertical, position) {
    const guide = document.createElement("div");
    guide.className = "pico-ds-guide " + (vertical ? "pico-ds-guide-v" : "pico-ds-guide-h");
    if (vertical) guide.style.left = position + "px";
    else guide.style.top = position + "px";
    parentEl.appendChild(guide);
  }

  function smartSnapPosition(parentEl, moving, x, y, width, height, excluded) {
    const threshold = 5;
    let bestX = { delta: threshold + 1, guide: null };
    let bestY = { delta: threshold + 1, guide: null };
    const movingX = [x, x + width / 2, x + width];
    const movingY = [y, y + height / 2, y + height];
    for (const candidate of directDesignWidgets(parentEl)) {
      const index = Number(candidate.dataset.index);
      if (candidate === moving || excluded.includes(index) || candidate.classList.contains("pico-ds-hidden")) continue;
      const cx = parseInt(candidate.style.left, 10) || 0;
      const cy = parseInt(candidate.style.top, 10) || 0;
      const cw = parseInt(candidate.style.width, 10) || candidate.offsetWidth;
      const ch = parseInt(candidate.style.height, 10) || candidate.offsetHeight;
      const targetX = [cx, cx + cw / 2, cx + cw];
      const targetY = [cy, cy + ch / 2, cy + ch];
      for (const source of movingX) {
        for (const target of targetX) {
          const delta = target - source;
          if (Math.abs(delta) < Math.abs(bestX.delta)) bestX = { delta, guide: target };
        }
      }
      for (const source of movingY) {
        for (const target of targetY) {
          const delta = target - source;
          if (Math.abs(delta) < Math.abs(bestY.delta)) bestY = { delta, guide: target };
        }
      }
    }
    clearDesignGuides(parentEl);
    if (Math.abs(bestX.delta) <= threshold) {
      x += bestX.delta;
      addDesignGuide(parentEl, true, bestX.guide);
    }
    if (Math.abs(bestY.delta) <= threshold) {
      y += bestY.delta;
      addDesignGuide(parentEl, false, bestY.guide);
    }
    return { x: Math.round(x), y: Math.round(y) };
  }

  function bindDesignInteract(item, index, surfaceId, hostEl) {
    item.tabIndex = 0;
    item.addEventListener("keydown", (ev) => {
      if (!["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown"].includes(ev.key)) return;
      ev.preventDefault();
      ev.stopPropagation();
      if (item.dataset.locked === "true") return;
      const step = ev.shiftKey ? 10 : 1;
      const parentEl = item.offsetParent || item.parentElement;
      const selected = (hostEl._picoDsSelection || [index]).filter((value) => {
        const node = hostEl.querySelector('[data-index="' + value + '"]');
        return node && node.offsetParent === parentEl && node.dataset.locked !== "true";
      });
      const dx = ev.key === "ArrowLeft" ? -step : ev.key === "ArrowRight" ? step : 0;
      const dy = ev.key === "ArrowUp" ? -step : ev.key === "ArrowDown" ? step : 0;
      const changes = [];
      for (const selectedIndex of selected) {
        const node = hostEl.querySelector('[data-index="' + selectedIndex + '"]');
        const width = parseInt(node.style.width, 10) || node.offsetWidth;
        const height = parseInt(node.style.height, 10) || node.offsetHeight;
        const maxX = Math.max(0, (parentEl ? parentEl.clientWidth : 10000) - width);
        const maxY = Math.max(0, (parentEl ? parentEl.clientHeight : 10000) - height);
        const x = Math.max(0, Math.min(maxX, (parseInt(node.style.left, 10) || 0) + dx));
        const y = Math.max(0, Math.min(maxY, (parseInt(node.style.top, 10) || 0) + dy));
        node.style.left = x + "px";
        node.style.top = y + "px";
        changes.push({ index: selectedIndex, x, y, width, height });
      }
      emit(surfaceId, "selection", hostEl._picoDsSelection || [index]);
      updateDesignSelection(hostEl, hostEl._picoDsSelection || [index]);
      if (changes.length > 1) emit(surfaceId, "layouts", changes);
      else if (changes.length === 1) emit(surfaceId, "layout", changes[0]);
    });

    item.addEventListener("pointerdown", (ev) => {
      if (ev.button !== 0) return;
      ev.preventDefault();
      ev.stopPropagation();

      const handle = ev.target && ev.target.dataset ? ev.target.dataset.handle : "";
      let selection = Array.from(hostEl._picoDsSelection || []);
      const additive = ev.ctrlKey || ev.metaKey || ev.shiftKey;
      if (additive) {
        if (selection.includes(index)) selection = selection.filter((value) => value !== index);
        else selection.push(index);
      } else if (!selection.includes(index)) {
        selection = [index];
      }
      if (selection.length === 0) {
        hostEl._picoDsSelection = [];
        updateDesignSelection(hostEl, []);
        emit(surfaceId, "selection", []);
        return;
      }
      if (!selection.includes(index)) {
        hostEl._picoDsSelection = selection;
        updateDesignSelection(hostEl, selection);
        emit(surfaceId, "selection", selection);
        return;
      }
      if (selection[selection.length - 1] !== index && !additive) {
        selection = selection.filter((value) => value !== index).concat(index);
      }
      hostEl._picoDsSelection = selection;
      updateDesignSelection(hostEl, selection);
      emit(surfaceId, "selection", selection);
      item.focus({ preventScroll: true });
      if (item.dataset.locked === "true") return;

      const parentEl = item.offsetParent || item.parentElement;
      const boundsW = parentEl ? parentEl.clientWidth : 10000;
      const boundsH = parentEl ? parentEl.clientHeight : 10000;

      let x = parseInt(item.style.left, 10) || 0;
      let y = parseInt(item.style.top, 10) || 0;
      let w = parseInt(item.style.width, 10) || item.offsetWidth;
      let h = parseInt(item.style.height, 10) || item.offsetHeight;
      const startX = ev.clientX;
      const startY = ev.clientY;
      const orig = { x, y, w, h };
      const group = handle ? [item] : selection.map((selectedIndex) =>
        hostEl.querySelector('[data-index="' + selectedIndex + '"]')
      ).filter((node) => node && node.offsetParent === parentEl && node.dataset.locked !== "true");
      const groupOriginal = group.map((node) => ({
        node,
        index: Number(node.dataset.index),
        x: parseInt(node.style.left, 10) || 0,
        y: parseInt(node.style.top, 10) || 0,
        width: parseInt(node.style.width, 10) || node.offsetWidth,
        height: parseInt(node.style.height, 10) || node.offsetHeight
      }));
      let moved = false;

      hostEl._picoDsDragging = true;
      try {
        item.setPointerCapture(ev.pointerId);
      } catch (_) {}

      const onMove = (mv) => {
        const dx = mv.clientX - startX;
        const dy = mv.clientY - startY;
        if (!moved && Math.abs(dx) < 2 && Math.abs(dy) < 2) return;
        moved = true;

        if (handle === "se") {
          w = Math.max(16, snap(orig.w + dx, 4));
          h = Math.max(16, snap(orig.h + dy, 4));
        } else if (handle === "sw") {
          w = Math.max(16, snap(orig.w - dx, 4));
          h = Math.max(16, snap(orig.h + dy, 4));
          x = snap(orig.x + (orig.w - w), 4);
        } else if (handle === "ne") {
          w = Math.max(16, snap(orig.w + dx, 4));
          h = Math.max(16, snap(orig.h - dy, 4));
          y = snap(orig.y + (orig.h - h), 4);
        } else if (handle === "nw") {
          w = Math.max(16, snap(orig.w - dx, 4));
          h = Math.max(16, snap(orig.h - dy, 4));
          x = snap(orig.x + (orig.w - w), 4);
          y = snap(orig.y + (orig.h - h), 4);
        } else {
          let deltaX = snap(dx, 4);
          let deltaY = snap(dy, 4);
          const minGroupX = Math.min(...groupOriginal.map((entry) => entry.x));
          const minGroupY = Math.min(...groupOriginal.map((entry) => entry.y));
          const maxGroupX = Math.max(...groupOriginal.map((entry) => entry.x + entry.width));
          const maxGroupY = Math.max(...groupOriginal.map((entry) => entry.y + entry.height));
          deltaX = Math.max(-minGroupX, Math.min(boundsW - maxGroupX, deltaX));
          deltaY = Math.max(-minGroupY, Math.min(boundsH - maxGroupY, deltaY));
          const snapped = smartSnapPosition(
            parentEl, item, orig.x + deltaX, orig.y + deltaY, orig.w, orig.h,
            groupOriginal.map((entry) => entry.index)
          );
          deltaX += snapped.x - (orig.x + deltaX);
          deltaY += snapped.y - (orig.y + deltaY);
          for (const entry of groupOriginal) {
            entry.node.style.left = entry.x + deltaX + "px";
            entry.node.style.top = entry.y + deltaY + "px";
          }
          x = orig.x + deltaX;
          y = orig.y + deltaY;
        }

        if (handle) {
          x = Math.max(0, Math.min(x, Math.max(0, boundsW - w)));
          y = Math.max(0, Math.min(y, Math.max(0, boundsH - h)));

          item.style.left = x + "px";
          item.style.top = y + "px";
          item.style.width = w + "px";
          item.style.height = h + "px";
        }
      };

      const onUp = () => {
        item.removeEventListener("pointermove", onMove);
        item.removeEventListener("pointerup", onUp);
        item.removeEventListener("pointercancel", onUp);
        hostEl._picoDsDragging = false;
        clearDesignGuides(parentEl);
        if (moved) {
          if (!handle && groupOriginal.length > 1) {
            emit(surfaceId, "layouts", groupOriginal.map((entry) => ({
              index: entry.index,
              x: parseInt(entry.node.style.left, 10) || 0,
              y: parseInt(entry.node.style.top, 10) || 0,
              width: entry.width,
              height: entry.height
            })));
          } else {
            emit(surfaceId, "layout", { index, x, y, width: w, height: h });
          }
        }
      };

      item.addEventListener("pointermove", onMove);
      item.addEventListener("pointerup", onUp);
      item.addEventListener("pointercancel", onUp);
    });
  }

  function renderDesignSurface(el, props, id) {
    if (el._picoDsDragging) {
      el._picoDsProps = Object.assign({}, el._picoDsProps || {}, props);
      return;
    }

    const prev = el._picoDsProps || {};
    const widgets = Array.isArray(props.widgets) ? props.widgets : [];
    const prevWidgets = Array.isArray(prev.widgets) ? prev.widgets : null;
    const geometrySame =
      prevWidgets &&
      JSON.stringify(prevWidgets) === JSON.stringify(widgets) &&
      prev.title === props.title &&
      prev.width === props.width &&
      prev.height === props.height;

    if (geometrySame && el.querySelector(".pico-ds-client")) {
      const selection = designSelection(props);
      const selected = selection.length ? selection[selection.length - 1] : -1;
      el._picoDsSelection = selection;
      el.querySelectorAll(".pico-ds-widget").forEach((node) => {
        const idx = Number(node.dataset.index);
        const on = selection.includes(idx);
        node.classList.toggle("pico-ds-selected", on);
        node.classList.toggle("pico-ds-primary", idx === selected);
        node.querySelectorAll(".pico-ds-handle").forEach((h) => h.remove());
        if (idx === selected && node.dataset.locked !== "true") addResizeHandles(node);
      });
      el._picoDsProps = Object.assign({}, props, { widgets });
      return;
    }

    clear(el);
    const title = props.title || "MyWindow";
    const width = Math.max(160, num(props.width, 480));
    const height = Math.max(120, num(props.height, 360));
    const selection = designSelection(props);
    const selected = selection.length ? selection[selection.length - 1] : -1;
    el._picoDsSelection = selection;

    const scroll = document.createElement("div");
    scroll.className = "pico-ds-scroll";

    const chrome = document.createElement("div");
    chrome.className = "pico-ds-chrome";
    chrome.style.width = width + "px";

    const titlebar = document.createElement("div");
    titlebar.className = "pico-ds-titlebar";
    const titleText = document.createElement("span");
    titleText.textContent = title;
    const closeBtn = document.createElement("span");
    closeBtn.className = "pico-ds-close";
    closeBtn.textContent = "✕";
    titlebar.appendChild(titleText);
    titlebar.appendChild(closeBtn);

    const client = document.createElement("div");
    client.className = "pico-ds-client";
    client.style.height = height + "px";
    client.addEventListener("pointerdown", (ev) => {
      if (ev.target === client) {
        el._picoDsSelection = [];
        updateDesignSelection(el, []);
        emit(id, "selection", []);
      }
    });

    const byParent = new Map();
    widgets.forEach((w, index) => {
      const parent = (w && w.parent) || "";
      if (!byParent.has(parent)) byParent.set(parent, []);
      byParent.get(parent).push({ w, index });
    });

    const bodies = new Map(); // widget id -> container body element
    bodies.set("", client);

    function mountLevel(parentKey) {
      const list = byParent.get(parentKey) || [];
      list.sort((a, b) => {
        const az = num(a.w.effectiveZIndex, isContainerKind(a.w.kind) ? num(a.w.zIndex, 0) : 1000 + num(a.w.zIndex, 0));
        const bz = num(b.w.effectiveZIndex, isContainerKind(b.w.kind) ? num(b.w.zIndex, 0) : 1000 + num(b.w.zIndex, 0));
        if (az !== bz) return az - bz;
        const ay = num(a.w.y, 0);
        const by = num(b.w.y, 0);
        if (ay !== by) return ay - by;
        return num(a.w.x, 0) - num(b.w.x, 0);
      });
      const host = bodies.get(parentKey) || client;
      for (const { w, index } of list) {
        const kind = (w && w.kind) || "label";
        const def = defaultWidgetSize(kind);
        const x = Math.max(0, num(w.x, 0));
        const y = Math.max(0, num(w.y, 0));
        const ww = Math.max(16, num(w.width, def.w));
        const hh = Math.max(16, num(w.height, def.h));

        const item = document.createElement("div");
        item.className = "pico-ds-widget pico-ds-" + kind;
        String(w.class || "").split(/\s+/).filter(Boolean).forEach((name) => item.classList.add(name));
        if (isContainerKind(kind)) item.classList.add("pico-ds-container");
        if (selection.includes(index)) item.classList.add("pico-ds-selected");
        if (selected === index) item.classList.add("pico-ds-primary");
        if (w && w.locked) item.classList.add("pico-ds-locked");
        if (w && w.hidden) item.classList.add("pico-ds-hidden");
        item.dataset.index = String(index);
        item.dataset.id = (w && w.id) || "";
        item.dataset.locked = w && w.locked ? "true" : "false";
        item.title = ((w && w.id) || kind) + "";
        item.style.left = x + "px";
        item.style.top = y + "px";
        item.style.width = ww + "px";
        item.style.height = hh + "px";
        item.style.zIndex = String(num(w.effectiveZIndex, isContainerKind(kind) ? num(w.zIndex, 0) : 1000 + num(w.zIndex, 0)));
        applyAppearance(item, w && w.appearance);

        const body = fillWidgetContent(item, w, kind);
        const badge = document.createElement("span");
        badge.className = "pico-ds-idtag";
        badge.textContent = ((w && w.locked) ? "🔒 " : "") + ((w && w.hidden) ? "◌ " : "") + ((w && w.id) || kind);
        item.appendChild(badge);
        if (selected === index && !(w && w.locked)) addResizeHandles(item);

        bindDesignInteract(item, index, id, el);
        host.appendChild(item);

        if (body && w && w.id) {
          bodies.set(w.id, body);
          mountLevel(w.id);
        }
      }
    }

    if (widgets.length === 0) {
      const empty = document.createElement("div");
      empty.className = "pico-ds-empty";
      empty.textContent = "Add controls from the toolbox · drag to move · handles to resize";
      client.appendChild(empty);
    } else {
      mountLevel("");
      // Orphans whose parent id is missing still show at root.
      widgets.forEach((w, index) => {
        if (!w || !w.parent) return;
        if (bodies.has(w.parent)) return;
        const already = client.querySelector('[data-index="' + index + '"]');
        if (already) return;
        const kind = w.kind || "label";
        const def = defaultWidgetSize(kind);
        const item = document.createElement("div");
        item.className = "pico-ds-widget pico-ds-" + kind;
        String(w.class || "").split(/\s+/).filter(Boolean).forEach((name) => item.classList.add(name));
        if (selection.includes(index)) item.classList.add("pico-ds-selected");
        if (selected === index) item.classList.add("pico-ds-primary");
        if (w.locked) item.classList.add("pico-ds-locked");
        if (w.hidden) item.classList.add("pico-ds-hidden");
        item.dataset.index = String(index);
        item.dataset.locked = w.locked ? "true" : "false";
        item.style.left = num(w.x, 0) + "px";
        item.style.top = num(w.y, 0) + "px";
        item.style.width = num(w.width, def.w) + "px";
        item.style.height = num(w.height, def.h) + "px";
        item.style.zIndex = String(num(w.effectiveZIndex, isContainerKind(kind) ? num(w.zIndex, 0) : 1000 + num(w.zIndex, 0)));
        applyAppearance(item, w && w.appearance);
        fillWidgetContent(item, w, kind);
        if (selected === index && !w.locked) addResizeHandles(item);
        bindDesignInteract(item, index, id, el);
        client.appendChild(item);
      });
    }

    chrome.appendChild(titlebar);
    chrome.appendChild(client);
    scroll.appendChild(chrome);
    el.appendChild(scroll);
    el._picoDsProps = Object.assign({}, props, { widgets });
  }

  function renderTreeNodes(container, list, treeId, selected) {
    if (!Array.isArray(list)) return;
    for (const node of list) {
      const item = document.createElement("div");
      item.className = "pico-tree-item";

      const row = document.createElement("div");
      row.className = "pico-tree-row";
      const hasChildren = Array.isArray(node.children) && node.children.length > 0;
      row.setAttribute("role", "treeitem");
      row.tabIndex = node.id === selected ? 0 : -1;
      row.setAttribute("aria-selected", node.id === selected ? "true" : "false");
      if (node.id === selected) row.classList.add("pico-selected");
      if (hasChildren) row.setAttribute("aria-expanded", node.expanded === false ? "false" : "true");

      const twisty = document.createElement("button");
      twisty.type = "button";
      twisty.className = "pico-tree-twisty";
      twisty.textContent = hasChildren ? (node.expanded === false ? "▸" : "▾") : "";
      twisty.disabled = !hasChildren;
      if (hasChildren) {
        twisty.addEventListener("click", (ev) => {
          ev.stopPropagation();
          emit(treeId, "toggle", { id: node.id, expanded: !(node.expanded !== false) });
        });
      }

      const label = document.createElement("span");
      label.className = "pico-tree-label";
      label.textContent = node.text || "";
      label.addEventListener("click", () => emit(treeId, "select", node.id));
      row.addEventListener("keydown", (ev) => {
        if (ev.key === "Enter" || ev.key === " ") {
          ev.preventDefault();
          emit(treeId, "select", node.id);
        } else if (hasChildren && (ev.key === "ArrowRight" || ev.key === "ArrowLeft")) {
          ev.preventDefault();
          emit(treeId, "toggle", { id: node.id, expanded: ev.key === "ArrowRight" });
        } else if (ev.key === "ArrowDown" || ev.key === "ArrowUp") {
          ev.preventDefault();
          const tree = container.closest(".pico-tree");
          const rows = tree ? Array.from(tree.querySelectorAll(".pico-tree-row")) : [];
          const index = rows.indexOf(row);
          const next = rows[index + (ev.key === "ArrowDown" ? 1 : -1)];
          if (next) next.focus();
        }
      });

      row.appendChild(twisty);
      row.appendChild(label);
      item.appendChild(row);

      if (hasChildren && node.expanded !== false) {
        const kids = document.createElement("div");
        kids.className = "pico-tree-children";
        renderTreeNodes(kids, node.children, treeId, selected);
        item.appendChild(kids);
      }
      container.appendChild(item);
    }
  }

  function setupTabs(el, id, selected) {
    const pages = Array.from(el.children).filter((child) => child.classList.contains("pico-tab-page"));
    const headers = document.createElement("div");
    headers.className = "pico-tab-headers";
    headers.setAttribute("role", "tablist");
    const body = document.createElement("div");
    body.className = "pico-tab-body";
    pages.forEach((page, index) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "pico-tab-header";
      button.textContent = page.dataset.title || ("Tab " + (index + 1));
      button.setAttribute("role", "tab");
      button.addEventListener("click", () => {
        updateTabs(el, index);
        emit(id, "change", index);
      });
      headers.appendChild(button);
      body.appendChild(page);
    });
    clear(el);
    el.appendChild(headers);
    el.appendChild(body);
    updateTabs(el, selected);
  }

  function updateTabs(el, selected) {
    const index = Math.max(0, Number(selected) || 0);
    const headers = Array.from(el.querySelectorAll(":scope > .pico-tab-headers > .pico-tab-header"));
    const pages = Array.from(el.querySelectorAll(":scope > .pico-tab-body > .pico-tab-page"));
    headers.forEach((header, i) => {
      header.classList.toggle("pico-selected", i === index);
      header.setAttribute("aria-selected", i === index ? "true" : "false");
      header.tabIndex = i === index ? 0 : -1;
    });
    pages.forEach((page, i) => {
      page.hidden = i !== index;
      page.setAttribute("aria-hidden", i === index ? "false" : "true");
    });
  }

  function createElement(node) {
    const props = node.props || {};
    let el;

    switch (node.kind) {
      case "column":
        el = document.createElement("div");
        el.className = "pico-column";
        break;
      case "row":
        el = document.createElement("div");
        el.className = "pico-row";
        break;
      case "stack":
        el = document.createElement("div");
        el.className = "pico-stack";
        break;
      case "grid":
        el = document.createElement("div");
        el.className = "pico-grid";
        el.style.gridTemplateColumns = `repeat(${props.columns || 1}, minmax(0, 1fr))`;
        break;
      case "split":
        el = document.createElement("div");
        el.className = props.vertical ? "pico-split pico-split-vertical" : "pico-split";
        el.style.setProperty("--pico-split-ratio", Math.max(10, Math.min(90, Number(props.ratio) || 50)) + "%");
        break;
      case "tabs":
        el = document.createElement("div");
        el.className = "pico-tabs";
        break;
      case "tab":
        el = document.createElement("section");
        el.className = "pico-tab-page";
        el.dataset.title = props.title || "";
        el.setAttribute("role", "tabpanel");
        break;
      case "dock":
        el = document.createElement("div");
        el.className = "pico-dock";
        break;
      case "dockitem":
        el = document.createElement("div");
        el.className = "pico-dock-item pico-dock-" + (props.region || "center");
        break;
      case "canvas":
        el = document.createElement("div");
        el.className = "pico-canvas";
        break;
      case "positioned":
        el = document.createElement("div");
        el.className = "pico-positioned";
        el.style.left = Number(props.x || 0) + "px";
        el.style.top = Number(props.y || 0) + "px";
        el.style.width = Math.max(1, Number(props.width) || 1) + "px";
        el.style.height = Math.max(1, Number(props.height) || 1) + "px";
        el.style.zIndex = String(Number(props.zIndex) || 0);
        break;
      case "form":
        el = document.createElement("div");
        el.className = "pico-form";
        break;
      case "field":
        el = document.createElement("div");
        el.className = "pico-field";
        break;
      case "label":
        el = document.createElement("div");
        el.className = "pico-label";
        el.textContent = props.text || "";
        break;
      case "button":
        el = document.createElement("button");
        el.className = "pico-button";
        el.type = "button";
        el.textContent = props.text || "";
        el.addEventListener("click", () => emit(node.id, "click"));
        break;
      case "textbox": {
        el = document.createElement("input");
        el.className = "pico-textbox";
        el.type = "text";
        el.value = props.value || "";
        el.addEventListener("input", () => {
          delete el._picoPendingValue;
          emit(node.id, "change", el.value);
        });
        el.addEventListener("blur", () => applyPendingValue(el));
        break;
      }
      case "numberbox": {
        el = document.createElement("input");
        el.className = "pico-numberbox";
        el.type = "number";
        el.value = props.value != null ? String(props.value) : "0";
        el.addEventListener("input", () => {
          delete el._picoPendingValue;
          const n = el.value === "" ? 0 : Number(el.value);
          emit(node.id, "change", Number.isFinite(n) ? n : 0);
        });
        el.addEventListener("blur", () => applyPendingValue(el));
        break;
      }
      case "checkbox": {
        el = document.createElement("label");
        el.className = "pico-checkbox";
        const input = document.createElement("input");
        input.type = "checkbox";
        input.checked = !!props.checked;
        input.addEventListener("change", () => emit(node.id, "change", input.checked));
        const span = document.createElement("span");
        span.className = "pico-checkbox-label";
        span.textContent = props.text || "";
        el.appendChild(input);
        el.appendChild(span);
        break;
      }
      case "combobox": {
        el = document.createElement("select");
        el.className = "pico-combobox";
        const items = Array.isArray(props.items) ? props.items : [];
        for (const item of items) {
          const opt = document.createElement("option");
          opt.value = item;
          opt.textContent = item;
          el.appendChild(opt);
        }
        el.value = props.value || (items[0] || "");
        el.addEventListener("change", () => {
          delete el._picoPendingValue;
          emit(node.id, "change", el.value);
        });
        el.addEventListener("blur", () => applyPendingValue(el));
        break;
      }
      case "table":
        el = document.createElement("div");
        el.className = "pico-table";
        renderTable(el, props, node.id);
        break;
      case "designsurface":
        el = document.createElement("div");
        el.className = "pico-designsurface";
        renderDesignSurface(el, props, node.id);
        break;
      case "tree":
        el = document.createElement("div");
        el.className = "pico-tree";
        el.setAttribute("role", "tree");
        renderTreeNodes(el, props.nodes || [], node.id, props.selected || "");
        break;
      case "dropzone": {
        el = document.createElement("div");
        el.className = "pico-dropzone";
        el.tabIndex = 0;
        el.setAttribute("role", "group");
        el.setAttribute("aria-label", props.prompt || "Drop files here");
        const prevent = (ev) => {
          ev.preventDefault();
          ev.stopPropagation();
        };
        el.addEventListener("dragenter", (ev) => {
          prevent(ev);
          el.classList.add("pico-drag-over");
        });
        el.addEventListener("dragover", prevent);
        el.addEventListener("dragleave", (ev) => {
          prevent(ev);
          if (!el.contains(ev.relatedTarget)) el.classList.remove("pico-drag-over");
        });
        el.addEventListener("drop", async (ev) => {
          prevent(ev);
          el.classList.remove("pico-drag-over");
          const maxBytes = Math.max(0, Number(props.maxBytes) || 0);
          const source = Array.from((ev.dataTransfer && ev.dataTransfer.files) || []);
          const files = await Promise.all(source.map(async (file) => {
            const item = {
              name: file.name,
              size: file.size,
              type: file.type,
              lastModified: file.lastModified,
              truncated: maxBytes > 0 && file.size > maxBytes
            };
            if (maxBytes > 0) {
              const blob = file.slice(0, maxBytes);
              item.data = await new Promise((resolve) => {
                const reader = new FileReader();
                reader.onload = () => resolve(String(reader.result || ""));
                reader.onerror = () => resolve("");
                reader.readAsDataURL(blob);
              });
            }
            return item;
          }));
          emit(node.id, "drop", files);
        });
        break;
      }
      default: {
        const plugins = window.__picoPlugins && window.__picoPlugins.controls;
        const spec = plugins && plugins[node.kind];
        if (spec && typeof spec.create === "function") {
          el = spec.create(node, props, { emit });
          if (!el) {
            el = document.createElement("div");
            el.className = "pico-plugin-error";
            el.textContent = "plugin create returned null: " + (node.kind || "");
          }
        } else {
          el = document.createElement("div");
          el.className = "pico-unknown";
          el.textContent = node.kind || "unknown";
        }
        break;
      }
    }

    el.dataset.picoId = node.id;
    applyProps(el, node.kind, props);

    if (Array.isArray(node.children)) {
      for (const child of node.children) {
        el.appendChild(createElement(child));
      }
    }
    if (node.kind === "tabs") setupTabs(el, node.id, props.selected);

    nodes.set(node.id, { el, kind: node.kind });
    return el;
  }

  function inputEl(el, kind) {
    if (kind === "checkbox") return el.querySelector("input");
    return el;
  }

  function applyPendingValue(control) {
    if (!control || !Object.prototype.hasOwnProperty.call(control, "_picoPendingValue")) return;
    const value = control._picoPendingValue;
    delete control._picoPendingValue;
    control.value = value != null ? String(value) : "";
  }

  function applyProps(el, kind, props) {
    const control = inputEl(el, kind);

    if (Object.prototype.hasOwnProperty.call(props, "class")) {
      const previous = (el.dataset.picoCustomClass || "").split(/\s+/).filter(Boolean);
      if (previous.length) el.classList.remove(...previous);
      const next = String(props.class || "").trim().split(/\s+/).filter(Boolean);
      if (next.length) el.classList.add(...next);
      el.dataset.picoCustomClass = next.join(" ");
    }

    if (Object.prototype.hasOwnProperty.call(props, "text")) {
      if (kind === "label" || kind === "button") el.textContent = props.text;
      else if (kind === "checkbox") {
        const span = el.querySelector(".pico-checkbox-label");
        if (span) span.textContent = props.text;
      } else if (kind === "field") {
        const lab = el.querySelector(".pico-field-label");
        if (lab) lab.textContent = props.text;
      }
    }

    if (Object.prototype.hasOwnProperty.call(props, "value")) {
      if (kind === "textbox" || kind === "numberbox" || kind === "combobox") {
        if (document.activeElement !== control) {
          control.value = props.value != null ? String(props.value) : "";
        } else {
          control._picoPendingValue = props.value;
        }
      }
    }

    if (Object.prototype.hasOwnProperty.call(props, "checked") && kind === "checkbox") {
      control.checked = !!props.checked;
    }

    if (Object.prototype.hasOwnProperty.call(props, "items") && kind === "combobox") {
      const current = control.value;
      clear(control);
      for (const item of props.items || []) {
        const opt = document.createElement("option");
        opt.value = item;
        opt.textContent = item;
        control.appendChild(opt);
      }
      control.value = props.value != null ? String(props.value) : current;
    }

    if (Object.prototype.hasOwnProperty.call(props, "appearance")) {
      applyAppearance(el, props.appearance);
    }

    if (kind === "positioned" && Object.prototype.hasOwnProperty.call(props, "zIndex")) {
      el.style.zIndex = String(Number(props.zIndex) || 0);
    }

    if (kind === "table" && (props.rows || props.columns || Object.prototype.hasOwnProperty.call(props, "selected"))) {
      renderTable(el, props, el.dataset.picoId);
    }

    if (kind === "designsurface" && (props.widgets || props.title || props.width || props.height || Object.prototype.hasOwnProperty.call(props, "selected") || Object.prototype.hasOwnProperty.call(props, "selection"))) {
      renderDesignSurface(el, props, el.dataset.picoId);
    }

    if (kind === "tree" && (props.nodes || Object.prototype.hasOwnProperty.call(props, "selected"))) {
      clear(el);
      renderTreeNodes(el, props.nodes || [], el.dataset.picoId, props.selected || "");
    }

    const plugins = window.__picoPlugins && window.__picoPlugins.controls;
    const pluginSpec = plugins && plugins[kind];
    if (pluginSpec && typeof pluginSpec.patch === "function") {
      pluginSpec.patch(el, props, { emit });
    }

    if (Object.prototype.hasOwnProperty.call(props, "columns") && kind === "grid") {
      el.style.gridTemplateColumns = `repeat(${props.columns || 1}, minmax(0, 1fr))`;
    }

    if (kind === "tabs" && Object.prototype.hasOwnProperty.call(props, "selected")) {
      updateTabs(el, props.selected);
    }

    if (Object.prototype.hasOwnProperty.call(props, "visible")) {
      el.classList.toggle("pico-hidden", props.visible === false);
    }

    if (Object.prototype.hasOwnProperty.call(props, "enabled")) {
      const disabled = props.enabled === false;
      if (kind === "checkbox") control.disabled = disabled;
      else if ("disabled" in control) control.disabled = disabled;
    }

    if (Object.prototype.hasOwnProperty.call(props, "animation") && props.animation) {
      const animation = props.animation;
      el.style.setProperty("--pico-animation-duration", Math.max(1, Number(animation.durationMS) || 180) + "ms");
      el.style.setProperty("--pico-animation-delay", Math.max(0, Number(animation.delayMS) || 0) + "ms");
      el.style.setProperty("--pico-animation-iterations", Math.max(1, Number(animation.iterations) || 1));
      el.classList.remove("pico-animate-fade-in", "pico-animate-slide-up", "pico-animate-scale-in", "pico-animate-pulse");
      void el.offsetWidth;
      if (animation.name) el.classList.add("pico-animate-" + animation.name);
    }
  }

  function mount(tree) {
    nodes.clear();
    clear(app);
    if (!tree) return;
    app.appendChild(createElement(tree));
  }

  function patch(payload) {
    if (!payload || !payload.id) return;
    const entry = nodes.get(payload.id);
    if (!entry) return;
    applyProps(entry.el, entry.kind, payload.props || {});
  }

  function handleMessage(raw) {
    let msg;
    try {
      msg = typeof raw === "string" ? JSON.parse(raw) : raw;
    } catch {
      return;
    }
    switch (msg.kind) {
      case "mount":
        mount(typeof msg.payload === "string" ? JSON.parse(msg.payload) : msg.payload);
        break;
      case "patch": {
        const p = typeof msg.payload === "string" ? JSON.parse(msg.payload) : msg.payload;
        patch(p);
        break;
      }
      case "call": {
        const p = typeof msg.payload === "string" ? JSON.parse(msg.payload) : msg.payload;
        if (msg.event === "theme") setTheme(p);
        else if (msg.event === "dialog.open") openDialog(p);
        else if (msg.event === "dialog.close") closeDialog();
        break;
      }
      case "request":
        handleRPC(msg);
        break;
      default:
        break;
    }
  }

  window.__picoReceive = function (raw) {
    handleMessage(raw);
  };

  if (Array.isArray(window.__picoPending)) {
    const pending = window.__picoPending.slice();
    window.__picoPending = [];
    for (const item of pending) handleMessage(item);
  }

  function signalReady(attempt) {
    if (post({ kind: "ready" })) return;
    if ((attempt || 0) > 200) return;
    setTimeout(() => signalReady((attempt || 0) + 1), 25);
  }

  let windowResizeTimer = 0;
  function reportWindowResize() {
    clearTimeout(windowResizeTimer);
    windowResizeTimer = setTimeout(() => emit("__pico_window__", "resize", {
      width: Math.round(window.innerWidth),
      height: Math.round(window.innerHeight)
    }), 50);
  }
  window.addEventListener("resize", reportWindowResize);
  setTimeout(reportWindowResize, 100);

  signalReady(0);
})();
