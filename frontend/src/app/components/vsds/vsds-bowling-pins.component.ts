import {
  Component, OnInit, OnDestroy, AfterViewInit,
  ViewChild, ElementRef, NgZone, ChangeDetectorRef,
} from '@angular/core';
import { FormsModule }          from '@angular/forms';
import { DecimalPipe }          from '@angular/common';
import * as THREE               from 'three';
import { OrbitControls }
  from 'three/examples/jsm/controls/OrbitControls.js';
import { HttpErrorResponse }    from '@angular/common/http';
import { AuthService }          from '../../auth/auth.service';
import {
  BundlesService,
  Bundle,
}                               from '../../services/bundles.service';
import {
  VsdsService,
  SectorVoxel,
}                               from '../../services/vsds.service';
import { galaxyMapUrl }         from '../../utils/galaxy-map';
import { fetchBundle }          from '../../utils/bundle-loader';
import {
  VIRIDIS, rhoLogT, tToRadius, rhoToHex,
  MIN_RADIUS, MAX_RADIUS,
}                               from '../../utils/color-scale';
import { SelectModule }         from 'primeng/select';
import { SelectButtonModule }   from 'primeng/selectbutton';
import { InputNumberModule }    from 'primeng/inputnumber';
import { ButtonModule }         from 'primeng/button';
import { MessageModule }        from 'primeng/message';

// ── Data types ─────────────────────────────────────────────────────────────

interface BundleSurvey {
  projectname: string;
  rho_max:     number;
  x:           number;
  z:           number;
  column_dev:  number | null;
  gc_x:        number;
  gc_z:        number;
  points:      { zsample: number; rho: number }[];
}

interface RhoScale {
  dataMin:  number;
  dataMax:  number;
  scaleMin: number;
  scaleMax: number;
  mode:     'global' | 'fit-to-view';
}

interface SelectOption {
  label: string;
  value: string;
}

// ── Default camera position ────────────────────────────────────────────────

// World coordinates are GC-relative (gc_x, gc_z). Galaxy in XZ plane.
//
// camera.up = (0,1,0) so OrbitControls rotates around world-Y.
// This means horizontal drag orbits in the XZ plane — "turning left/right".
//
// The camera must NOT sit on the Y axis (0,Y,0) with up=(0,1,0): that puts
// it at the OrbitControls north-pole singularity, where horizontal drag
// produces a roll instead of a turn.
//
// Default: elevated at ~59° above the galactic plane, looking from +Z.
//   r ≈ 116 620 ly, phi ≈ 31° from pole — responsive horizontal orbit.
//   FOV=45° → visible radius ≈ 48 300 ly, covers the ±45 000 ly image.
const CAM_DEFAULT_POS    = new THREE.Vector3(0, 100000, 60000);
const CAM_DEFAULT_TARGET = new THREE.Vector3(0, 0, 0);

@Component({
  selector:    'app-vsds-bowling-pins',
  standalone:  true,
  imports: [
    FormsModule,
    DecimalPipe,
    SelectModule,
    SelectButtonModule,
    InputNumberModule,
    ButtonModule,
    MessageModule,
  ],
  templateUrl: './vsds-bowling-pins.component.html',
  styleUrl:    './vsds-bowling-pins.component.scss',
})
export class VsdsBowlingPinsComponent
  implements OnInit, AfterViewInit, OnDestroy {

  @ViewChild('canvas') canvasRef!: ElementRef<HTMLCanvasElement>;

  // ── Source selection ───────────────────────────────────────────────────
  sourceOptions: SelectOption[] = [
    { label: 'Bundles',   value: 'bundles' },
    { label: 'Database',  value: 'db' },
  ];
  selectedSource  = 'bundles';
  bundles:         Bundle[] = [];
  filteredBundles: Bundle[] = [];
  selectedBundle:  Bundle | null = null;

  // ── DB source parameters ───────────────────────────────────────────────
  xzStep    = 50;
  yStep     = 20;
  dbLoading = false;
  dbError:  string | null = null;

  // ── State ──────────────────────────────────────────────────────────────
  loading   = false;
  loadError: string | null = null;

  // ── Scale ──────────────────────────────────────────────────────────────
  scale:     RhoScale | null = null;
  sliderPos  = 100;   // 0–100, log-mapped to [dataMin, dataMax]
  rhoUnit:   'kly3' | 'ly3' = 'kly3';
  scaleMode: 'global' | 'fit-to-view' = 'global';

  scaleModeOptions: SelectOption[] = [
    { label: 'Global',       value: 'global' },
    { label: 'Fit to view',  value: 'fit-to-view' },
  ];

  unitOptions: SelectOption[] = [
    { label: 'sys/kLy³',  value: 'kly3' },
    { label: 'sys/ly³',   value: 'ly3' },
  ];

  // ── Tooltip ────────────────────────────────────────────────────────────
  tooltipVisible = false;
  tooltipX       = 0;
  tooltipY       = 0;
  tooltipSurvey: BundleSurvey | null = null;

  // ── Three.js (private) ─────────────────────────────────────────────────
  private renderer!:    THREE.WebGLRenderer;
  private scene!:       THREE.Scene;
  private camera!:      THREE.PerspectiveCamera;
  private controls!:    OrbitControls;
  private animFrameId   = 0;
  private surveys:      BundleSurvey[] = [];
  private hitMeshes:    THREE.Mesh[]   = [];
  private pinGroups:    THREE.Group[]  = [];
  private galaxyPlane:  THREE.Mesh | null = null;
  private raycaster     = new THREE.Raycaster();
  private mouse         = new THREE.Vector2();
  private frustum       = new THREE.Frustum();
  private projMatrix    = new THREE.Matrix4();
  private sceneReady    = false;
  private bundleBaseUrl: string | null = null;
  private resizeObserver?: ResizeObserver;
  private fitTimer: ReturnType<typeof setTimeout> | null = null;
  private markerObjects: THREE.Object3D[] = [];

  // ── Event handlers (kept for removeEventListener) ─────────────────────
  private boundMouseMove?: (e: MouseEvent) => void;

  constructor(
    public  authService:    AuthService,
    private bundlesService: BundlesService,
    private vsdsService:    VsdsService,
    private ngZone:         NgZone,
    private cdr:            ChangeDetectorRef,
  ) {}

  // ── Lifecycle ──────────────────────────────────────────────────────────

  ngOnInit(): void {
    this.bundleBaseUrl = this.bundlesService.bundleBaseUrl;
    if (!this.bundleBaseUrl) {
      this.bundlesService.getConfig().subscribe({
        next: cfg => {
          this.bundleBaseUrl = cfg.bundleBaseUrl;
          if (this.sceneReady) this.createGalaxyPlane();
        },
      });
    }
    this.bundlesService.listBundles().subscribe({
      next: bundles => {
        this.bundles         = bundles;
        this.filteredBundles = bundles.filter(b =>
          b.measurementtype === 'vsds' &&
          b.status          === 'ready' &&
          b.subtype         === 'surveys',
        );
      },
    });
  }

  ngAfterViewInit(): void {
    this.ngZone.runOutsideAngular(() => {
      this.initScene();
      this.sceneReady = true;
      if (this.bundleBaseUrl) this.createGalaxyPlane();
      this.startLoop();
    });
  }

  ngOnDestroy(): void {
    if (this.animFrameId) cancelAnimationFrame(this.animFrameId);
    if (this.fitTimer !== null) clearTimeout(this.fitTimer);
    this.resizeObserver?.disconnect();
    const canvas = this.canvasRef?.nativeElement;
    if (canvas && this.boundMouseMove) {
      canvas.removeEventListener('mousemove', this.boundMouseMove);
    }
    this.clearPins();
    if (this.galaxyPlane) {
      this.scene?.remove(this.galaxyPlane);
      this.galaxyPlane.geometry.dispose();
      (this.galaxyPlane.material as THREE.Material).dispose();
    }
    this.controls?.dispose();
    this.renderer?.dispose();
  }

  // ── Scene init ─────────────────────────────────────────────────────────

  private initScene(): void {
    const canvas  = this.canvasRef.nativeElement;
    const wrapper = canvas.parentElement!;
    const w       = wrapper.clientWidth  || 800;
    const h       = wrapper.clientHeight || 600;

    this.renderer = new THREE.WebGLRenderer({
      canvas, antialias: true,
    });
    this.renderer.setSize(w, h);
    this.renderer.setPixelRatio(window.devicePixelRatio);
    this.renderer.toneMapping         = THREE.ACESFilmicToneMapping;
    this.renderer.toneMappingExposure = 0.9;

    this.camera = new THREE.PerspectiveCamera(45, w / h, 1, 200_000);
    this.camera.up.set(0, 1, 0);  // world-Y up: horizontal drag = turn, not roll
    this.camera.position.copy(CAM_DEFAULT_POS);
    this.camera.lookAt(CAM_DEFAULT_TARGET);

    this.scene = new THREE.Scene();
    this.scene.background = new THREE.Color(0x000000);
    this.scene.add(new THREE.AmbientLight(0xffffff, 0.6));
    const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
    dirLight.position.set(0, 10000, 0);
    this.scene.add(dirLight);

    this.controls = new OrbitControls(this.camera, canvas);
    // Prevent camera from reaching the Y-axis pole (phi=0) where horizontal
    // drag degenerates into a roll. maxPolarAngle=π allows looking from below.
    this.controls.minPolarAngle = 0.05;
    this.controls.maxPolarAngle = Math.PI;
    this.controls.dampingFactor = 0.05;
    this.controls.enableDamping = true;
    this.controls.target.copy(CAM_DEFAULT_TARGET);
    this.controls.addEventListener('change', () => {
      this.debouncedFrustumUpdate();
    });

    this.resizeObserver = new ResizeObserver(() => this.onResize());
    this.resizeObserver.observe(wrapper);

    this.boundMouseMove = (e: MouseEvent) => this.onMouseMove(e);
    canvas.addEventListener('mousemove', this.boundMouseMove);
  }

  private createGalaxyPlane(): void {
    if (!this.bundleBaseUrl || !this.scene) return;
    if (this.galaxyPlane) {
      this.scene.remove(this.galaxyPlane);
      this.galaxyPlane.geometry.dispose();
      (this.galaxyPlane.material as THREE.Material).dispose();
      this.galaxyPlane = null;
    }
    const url    = galaxyMapUrl(this.bundleBaseUrl);
    const loader = new THREE.TextureLoader();
    loader.load(url, (texture) => {
      texture.colorSpace = THREE.SRGBColorSpace;
      // flipY=false: UV V matches the negated-Z world axis.
      // X mirror is handled via scale.x = −1 on the mesh.
      texture.flipY = false;
			texture.rotation = Math.PI;
			texture.center = new THREE.Vector2(0.5, 0.5);
      const geo = new THREE.PlaneGeometry(90000, 90000);
      const mat = new THREE.MeshBasicMaterial({
        map:        texture,
        side:       THREE.DoubleSide,
        depthTest:  false,
        depthWrite: false,
      });
      this.galaxyPlane = new THREE.Mesh(geo, mat);
      this.galaxyPlane.renderOrder = -1;
      this.galaxyPlane.rotation.x = -Math.PI / 2;
      // Flip mesh in X so the spiral is counter-clockwise.
      // All world X positions use +gc_x to stay aligned.
      this.galaxyPlane.scale.x = -1;
      // scale.x=−1 mirrors the mesh in X → CCW spiral.
      // Plane center: gc_x=−25.21875 → world X=−25.21875.
      // All world positions use (+gc_x, 0, −gc_z).
      this.galaxyPlane.position.set(
        -25.21875, -1, 899.96875,
      );
      this.scene.add(this.galaxyPlane);
    });
  }

  private startLoop(): void {
    const loop = () => {
      this.animFrameId = requestAnimationFrame(loop);
      this.controls.update();
      this.renderer.render(this.scene, this.camera);
    };
    loop();
  }

  private onResize(): void {
    const wrapper = this.canvasRef.nativeElement.parentElement;
    if (!wrapper) return;
    const w = wrapper.clientWidth;
    const h = wrapper.clientHeight;
    if (w === 0 || h === 0) return;
    this.camera.aspect = w / h;
    this.camera.updateProjectionMatrix();
    this.renderer.setSize(w, h);
  }

  // ── Bundle loading ─────────────────────────────────────────────────────

  async onBundleChange(): Promise<void> {
    if (!this.selectedBundle || !this.bundleBaseUrl) return;
    this.loading   = true;
    this.loadError = null;
    this.clearPins();
    this.surveys = [];
    this.scale   = null;
    try {
      const url = this.bundleBaseUrl + this.selectedBundle.filename;
      const data = await fetchBundle<BundleSurvey>(url);
      this.surveys = data;
      this.initScale(data);
      this.ngZone.runOutsideAngular(() => this.buildPins());
    } catch (err) {
      this.loadError = err instanceof Error
        ? err.message : 'Failed to load bundle';
    } finally {
      this.loading = false;
      this.cdr.detectChanges();
    }
  }

  // ── DB loading ─────────────────────────────────────────────────────────

  loadFromDB(): void {
    this.dbLoading = true;
    this.dbError   = null;
    this.clearPins();
    this.surveys = [];
    this.scale   = null;
    this.vsdsService.listSectors(this.xzStep, this.yStep)
      .subscribe({
        next: (voxels: SectorVoxel[]) => {
          const surveyMap = new Map<string, BundleSurvey>();
          for (const v of voxels) {
            const key = `${v.gc_x}_${v.gc_z}`;
            if (!surveyMap.has(key)) {
              surveyMap.set(key, {
                projectname: 'DB Sectors',
                rho_max:     0,
                x:           v.gc_x + 25.21875,
                z:           v.gc_z + 25899.96875,
                column_dev:  null,
                gc_x:        v.gc_x,
                gc_z:        v.gc_z,
                points:      [],
              });
            }
            const s = surveyMap.get(key)!;
            s.points.push({
              zsample: (v.y_min + v.y_max) / 2,
              rho:     v.rho_avg,
            });
            if (v.rho_max > s.rho_max) s.rho_max = v.rho_max;
          }
          this.surveys = [...surveyMap.values()];
          this.initScale(this.surveys);
          this.ngZone.runOutsideAngular(() => this.buildPins());
          this.dbLoading = false;
          this.cdr.detectChanges();
        },
        error: (err: HttpErrorResponse) => {
          this.dbError   =
            err.error?.message ?? 'Failed to load sectors';
          this.dbLoading = false;
        },
      });
  }

  // ── Scale ──────────────────────────────────────────────────────────────

  private initScale(surveys: BundleSurvey[]): void {
    let dataMin =  Infinity;
    let dataMax = -Infinity;
    for (const s of surveys) {
      for (const p of s.points) {
        if (p.rho > 0 && p.rho < dataMin) dataMin = p.rho;
        if (p.rho > dataMax)              dataMax = p.rho;
      }
    }
    if (!isFinite(dataMin)) dataMin = 1e-6;
    if (!isFinite(dataMax)) dataMax = 1;
    this.scale = {
      dataMin, dataMax,
      scaleMin: dataMin,
      scaleMax: dataMax,
      mode: 'global',
    };
    this.scaleMode = 'global';
    this.sliderPos = 100;
  }

  resetScaleMax(): void {
    if (!this.scale || this.scaleMode !== 'global') return;
    this.scale     = { ...this.scale, scaleMax: this.scale.dataMax };
    this.sliderPos = 100;
    this.ngZone.runOutsideAngular(() => this.rebuildPins());
  }

  onScaleModeChange(): void {
    if (!this.scale) return;
    this.scale = { ...this.scale, mode: this.scaleMode };
    if (this.scaleMode === 'global') {
      this.scale     = { ...this.scale, scaleMax: this.scale.dataMax };
      this.sliderPos = this.rhoToSlider(this.scale.dataMax);
      this.ngZone.runOutsideAngular(() => this.rebuildPins());
    } else {
      this.debouncedFrustumUpdate();
    }
  }

  onSliderInput(event: Event): void {
    if (!this.scale || this.scaleMode !== 'global') return;
    const pos     = +(event.target as HTMLInputElement).value;
    const rawRho  = this.sliderToRho(pos);
    const snapped = this.snapToRound(rawRho);
    this.scale    = { ...this.scale, scaleMax: snapped };
    // Snap the slider thumb back to the rounded rho position.
    // To make this freeform (no snapping): remove snapToRound() call above
    // and use rawRho directly; then remove the sliderPos reassignment below.
    this.sliderPos = this.rhoToSlider(snapped);
    this.ngZone.runOutsideAngular(() => this.rebuildPins());
  }

  /**
   * Snap a positive value to 2 significant figures.
   * E.g. 0.001523 → 0.0015, 2.374 → 2.4, 13.7 → 14.
   *
   * To disable snapping and allow freeform values: remove the call to
   * snapToRound() in onSliderInput() and use rawRho directly.
   */
  private snapToRound(value: number): number {
    if (value <= 0) return value;
    const mag   = Math.floor(Math.log10(value));
    const scale = Math.pow(10, mag - 1);   // 2 significant figures
    return Math.round(value / scale) * scale;
  }

  /** Map slider position [0, 100] → raw rho on a log scale. */
  private sliderToRho(pos: number): number {
    const logMin = Math.log(this.scale!.dataMin);
    const logMax = Math.log(this.scale!.dataMax);
    return Math.exp(logMin + (pos / 100) * (logMax - logMin));
  }

  /** Map raw rho → slider position [0, 100] on a log scale. */
  private rhoToSlider(rho: number): number {
    const logMin = Math.log(this.scale!.dataMin);
    const logMax = Math.log(this.scale!.dataMax);
    if (logMax <= logMin) return 100;
    const t = (Math.log(rho) - logMin) / (logMax - logMin);
    return Math.max(0, Math.min(100, t * 100));
  }

  // ── Fit-to-view ────────────────────────────────────────────────────────

  private debouncedFrustumUpdate(): void {
    if (this.fitTimer !== null) clearTimeout(this.fitTimer);
    this.fitTimer = setTimeout(() => {
      this.fitTimer = null;
      if (this.scaleMode === 'fit-to-view') this.updateFitToView();
    }, 100);
  }

  private updateFitToView(): void {
    if (!this.scale || this.surveys.length === 0) return;
    this.camera.updateMatrixWorld();
    this.projMatrix.multiplyMatrices(
      this.camera.projectionMatrix,
      this.camera.matrixWorldInverse,
    );
    this.frustum.setFromProjectionMatrix(this.projMatrix);
    const tmp = new THREE.Vector3();
    const visible = this.surveys.filter(s => {
      tmp.set(s.gc_x, 0, -s.gc_z);
      return this.frustum.containsPoint(tmp);
    });
    if (visible.length === 0) return;
    let viewMax = -Infinity;
    let viewMin =  Infinity;
    for (const s of visible) {
      if (s.rho_max > viewMax) viewMax = s.rho_max;
      if (s.rho_max < viewMin) viewMin = s.rho_max;
    }
    if (!isFinite(viewMax) || !isFinite(viewMin)) return;
    this.scale = {
      ...this.scale,
      scaleMin: viewMin,
      scaleMax: viewMax,
      mode: 'fit-to-view',
    };
    this.ngZone.run(() => {
      this.sliderPos = this.rhoToSlider(viewMax);
      this.cdr.detectChanges();
    });
    this.rebuildPins();
  }

  // ── Pin building ───────────────────────────────────────────────────────

  private clearPins(): void {
    for (const g of this.pinGroups) {
      this.scene?.remove(g);
      g.traverse(obj => {
        if ((obj as THREE.Mesh).isMesh) {
          const m = obj as THREE.Mesh;
          m.geometry.dispose();
          (m.material as THREE.Material).dispose();
        }
      });
    }
    for (const h of this.hitMeshes) {
      this.scene?.remove(h);
      h.geometry.dispose();
      (h.material as THREE.Material).dispose();
    }
    this.pinGroups = [];
    this.hitMeshes = [];
  }

  private buildPins(): void {
    if (!this.scale || !this.scene) return;
    this.clearPins();
    const { scaleMin, scaleMax } = this.scale;
    for (let si = 0; si < this.surveys.length; si++) {
      const s   = this.surveys[si];
      const pts =
        [...s.points].sort((a, b) => a.zsample - b.zsample);
      if (pts.length === 0) continue;

      const group = new THREE.Group();
      // World convention: (+gc_x, 0, −gc_z).
      group.position.set(s.gc_x, 0, -s.gc_z);
      this.buildPin(group, pts, scaleMin, scaleMax);
      this.scene.add(group);
      this.pinGroups.push(group);

      const first  = pts[0];
      const last   = pts[pts.length - 1];
      const y0     = first.zsample - 50;
      const yTop   = last.zsample  + 50;
      const pinH   = Math.max(50, yTop - y0);
      const pinMid = (y0 + yTop) / 2;
      const hitGeo = new THREE.CylinderGeometry(
        MAX_RADIUS, MAX_RADIUS, pinH, 8,
      );
      const hitMat =
        new THREE.MeshBasicMaterial({ visible: false });
      const hit    = new THREE.Mesh(hitGeo, hitMat);
      hit.position.set(s.gc_x, pinMid, -s.gc_z);
      hit.userData['surveyIndex'] = si;
      this.scene.add(hit);
      this.hitMeshes.push(hit);
    }
  }

  // Build one smooth pin as a single LatheGeometry.
  // Profile is smoothed with a Catmull-Rom spline; vertex
  // colours encode the density at each height.
  private buildPin(
    group:    THREE.Group,
    pts:      { zsample: number; rho: number }[],
    scaleMin: number,
    scaleMax: number,
  ): void {
    const NRADIAL = 16;
    const NSUB    = 6; // spline sub-samples per data interval

    // Control points: cap → data → cap
    const raw: THREE.Vector2[] = [
      new THREE.Vector2(
        MIN_RADIUS, pts[0].zsample - 50,
      ),
    ];
    for (const p of pts) {
      raw.push(new THREE.Vector2(
        tToRadius(rhoLogT(p.rho, scaleMin, scaleMax)),
        p.zsample,
      ));
    }
    raw.push(new THREE.Vector2(
      MIN_RADIUS, pts[pts.length - 1].zsample + 50,
    ));

    // Smooth via Catmull-Rom; clamp radius to MIN_RADIUS
    const spline  = new THREE.SplineCurve(raw);
    const nOut    = (raw.length - 1) * NSUB + 1;
    const profile = spline.getPoints(nOut).map(p =>
      new THREE.Vector2(Math.max(MIN_RADIUS, p.x), p.y),
    );

    const geo = new THREE.LatheGeometry(profile, NRADIAL);

    // Assign vertex colours by reading the actual y from the
    // position buffer — no assumptions about internal layout.
    const pos    =
      geo.attributes['position'] as THREE.BufferAttribute;
    const nVerts = pos.count;
    const cBuf   = new Float32Array(nVerts * 3);
    for (let v = 0; v < nVerts; v++) {
      const rho = this.rhoAtY(pos.getY(v), pts);
      const hex = rhoToHex(rho, scaleMin, scaleMax, VIRIDIS);
      cBuf[v * 3    ] = ((hex >> 16) & 0xff) / 255;
      cBuf[v * 3 + 1] = ((hex >>  8) & 0xff) / 255;
      cBuf[v * 3 + 2] = ( hex        & 0xff) / 255;
    }
    geo.setAttribute(
      'color', new THREE.BufferAttribute(cBuf, 3),
    );

    group.add(new THREE.Mesh(geo,
      new THREE.MeshLambertMaterial({ vertexColors: true }),
    ));
  }

  // Linear interpolation of rho at a given y between sample pts.
  private rhoAtY(
    y:   number,
    pts: { zsample: number; rho: number }[],
  ): number {
    if (y <= pts[0].zsample) return pts[0].rho;
    const last = pts[pts.length - 1];
    if (y >= last.zsample)   return last.rho;
    for (let i = 0; i + 1 < pts.length; i++) {
      if (y <= pts[i + 1].zsample) {
        const span = pts[i + 1].zsample - pts[i].zsample;
        if (span === 0) return pts[i].rho;
        const t = (y - pts[i].zsample) / span;
        return pts[i].rho + t * (pts[i + 1].rho - pts[i].rho);
      }
    }
    return last.rho;
  }

  private rebuildPins(): void {
    if (!this.scale) return;
    this.buildPins();
  }

  // ── Camera ─────────────────────────────────────────────────────────────

  resetCamera(): void {
    this.camera.up.set(0, 1, 0);
    this.camera.position.copy(CAM_DEFAULT_POS);
    this.camera.lookAt(CAM_DEFAULT_TARGET);
    this.controls.target.copy(CAM_DEFAULT_TARGET);
    this.controls.update();
  }

  // ── Mouse hover ────────────────────────────────────────────────────────

  private onMouseMove(event: MouseEvent): void {
    const canvas = this.canvasRef.nativeElement;
    const rect   = canvas.getBoundingClientRect();
    this.mouse.set(
      ((event.clientX - rect.left) / rect.width)  * 2 - 1,
      -((event.clientY - rect.top)  / rect.height) * 2 + 1,
    );
    this.raycaster.setFromCamera(this.mouse, this.camera);
    const hits = this.raycaster.intersectObjects(this.hitMeshes);
    if (hits.length > 0) {
      const idx    = hits[0].object.userData['surveyIndex'] as number;
      const survey = this.surveys[idx];
      this.ngZone.run(() => {
        this.tooltipSurvey  = survey;
        this.tooltipX       = event.clientX;
        this.tooltipY       = event.clientY;
        this.tooltipVisible = true;
        this.cdr.detectChanges();
      });
    } else if (this.tooltipVisible) {
      this.ngZone.run(() => {
        this.tooltipVisible = false;
        this.cdr.detectChanges();
      });
    }
  }

  // ── Display helpers ────────────────────────────────────────────────────

  get unitLabel(): string {
    return this.rhoUnit === 'kly3' ? 'sys/kLy³' : 'sys/ly³';
  }

  /** Raw rho → display value in active unit. */
  toDisplay(rho: number): number {
    return this.rhoUnit === 'kly3' ? rho * 1000 : rho;
  }

  /** Format rho for display in active unit. */
  formatRho(rho: number | null | undefined): string {
    if (rho == null) return '—';
    const v = this.toDisplay(rho);
    if (v >= 100)  return v.toFixed(1);
    if (v >= 1)    return v.toFixed(2);
    if (v >= 0.01) return v.toFixed(4);
    return v.toExponential(2);
  }
}
