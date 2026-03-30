const BYTES_PER_MEGABYTE = 1024 * 1024;
const RECOMMENDED_LISTING_IMAGE_RATIO = 4 / 3;
export const RECOMMENDED_LISTING_IMAGE_RATIO_LABEL = "4:3";
const IMAGE_RATIO_TOLERANCE = 0.08;

export const MAX_LISTING_VIDEO_BYTES = 100 * BYTES_PER_MEGABYTE;
export const MAX_LISTING_VIDEO_DURATION_SECONDS = 60;

type ListingVideoPrecheckOptions = {
  readDuration?: (file: File) => Promise<number | null>;
};

type ListingImagePrecheckOptions = {
  readDimensions?: (file: File) => Promise<{ width: number; height: number } | null>;
};

export type ListingVideoPrecheckResult = {
  ok: boolean;
  durationSeconds: number | null;
  message: string;
};

export type ListingImagePrecheckResult = {
  message: string;
  offRatioCount: number;
};

export async function validateListingVideoSelection(
  file: File,
  options: ListingVideoPrecheckOptions = {},
): Promise<ListingVideoPrecheckResult> {
  if (file.size > MAX_LISTING_VIDEO_BYTES) {
    return {
      ok: false,
      durationSeconds: null,
      message: `Choose a video under ${formatVideoBytes(MAX_LISTING_VIDEO_BYTES)} before uploading. Backend validation still decides the final result.`,
    };
  }

  const readDuration = options.readDuration ?? readListingVideoDurationSeconds;
  const durationSeconds = await readDuration(file);

  if (durationSeconds != null && durationSeconds > MAX_LISTING_VIDEO_DURATION_SECONDS) {
    return {
      ok: false,
      durationSeconds,
      message: `Choose a video under ${formatDuration(MAX_LISTING_VIDEO_DURATION_SECONDS)} before uploading. Backend validation still decides the final result.`,
    };
  }

  return {
    ok: true,
    durationSeconds,
    message: buildVideoReadyMessage(file, durationSeconds),
  };
}

export async function inspectListingImageSelection(
  files: File[],
  options: ListingImagePrecheckOptions = {},
): Promise<ListingImagePrecheckResult> {
  if (files.length === 0) {
    return {
      message: describeSelectedImageFiles(files),
      offRatioCount: 0,
    };
  }

  const readDimensions = options.readDimensions ?? readListingImageDimensions;
  let offRatioCount = 0;

  for (const file of files) {
    const dimensions = await readDimensions(file);
    if (!dimensions || dimensions.height <= 0) {
      continue;
    }

    const ratio = dimensions.width / dimensions.height;
    if (Math.abs(ratio - RECOMMENDED_LISTING_IMAGE_RATIO) > IMAGE_RATIO_TOLERANCE) {
      offRatioCount += 1;
    }
  }

  if (offRatioCount === 0) {
    return {
      message: `${describeSelectedImageFiles(files)}. Recommended ratio ${RECOMMENDED_LISTING_IMAGE_RATIO_LABEL} looks good for listings cards.`,
      offRatioCount,
    };
  }

  return {
    message: `${describeSelectedImageFiles(files)}. Recommended ratio is ${RECOMMENDED_LISTING_IMAGE_RATIO_LABEL}; ${offRatioCount} image${offRatioCount > 1 ? "s differ" : " differs"} and will show with padding so nothing gets cropped in listings cards.`,
    offRatioCount,
  };
}

export function describeSelectedImageFiles(files: File[]) {
  if (files.length === 0) {
    return "No images selected yet";
  }

  if (files.length === 1) {
    return `Ready: ${files[0].name}`;
  }

  const preview = files
    .slice(0, 2)
    .map((file) => file.name)
    .join(", ");

  return `Ready: ${files.length} images selected${preview ? ` (${preview}${files.length > 2 ? ", ..." : ""})` : ""}`;
}

export function describeExistingListingVideo(fileName: string | null | undefined, durationSeconds?: number | null) {
  const parts = [fileName ?? "Listing video"];

  if (durationSeconds != null) {
    parts.push(formatDuration(durationSeconds));
  }

  return parts.join(" · ");
}

async function readListingImageDimensions(file: File): Promise<{ width: number; height: number } | null> {
  const objectUrl = URL.createObjectURL(file);

  try {
    return await new Promise<{ width: number; height: number } | null>((resolve, reject) => {
      const image = new Image();

      const cleanup = () => {
        image.src = "";
        URL.revokeObjectURL(objectUrl);
      };

      image.onload = () => {
        const width = image.naturalWidth;
        const height = image.naturalHeight;
        cleanup();
        if (width > 0 && height > 0) {
          resolve({ width, height });
          return;
        }
        resolve(null);
      };
      image.onerror = () => {
        cleanup();
        reject(new Error("We could not inspect that image file."));
      };

      image.src = objectUrl;
    });
  } catch {
    return null;
  }
}

export function formatVideoBytes(bytes: number) {
  const megabytes = bytes / BYTES_PER_MEGABYTE;

  if (Number.isInteger(megabytes)) {
    return `${megabytes} MB`;
  }

  return `${megabytes.toFixed(1)} MB`;
}

export function formatDuration(seconds: number) {
  const normalized = Math.max(0, Math.ceil(seconds));

  if (normalized < 60) {
    return `${normalized}s`;
  }

  const minutes = Math.floor(normalized / 60);
  const remainder = normalized % 60;

  return remainder === 0 ? `${minutes}m` : `${minutes}m ${remainder}s`;
}

export async function readListingVideoDurationSeconds(file: File): Promise<number | null> {
  const objectUrl = URL.createObjectURL(file);

  try {
    return await new Promise<number | null>((resolve, reject) => {
      const video = document.createElement("video");

      const cleanup = () => {
        video.removeAttribute("src");
        video.load();
        URL.revokeObjectURL(objectUrl);
      };

      video.preload = "metadata";
      video.onloadedmetadata = () => {
        const duration = Number.isFinite(video.duration) ? Math.ceil(video.duration) : null;

        cleanup();
        resolve(duration);
      };
      video.onerror = () => {
        cleanup();
        reject(new Error("We could not inspect that video file."));
      };
      video.src = objectUrl;
    });
  } catch {
    return null;
  }
}

function buildVideoReadyMessage(file: File, durationSeconds: number | null) {
  const details = [file.name, formatVideoBytes(file.size)];

  if (durationSeconds != null) {
    details.push(formatDuration(durationSeconds));
  } else {
    details.push("duration hint unavailable");
  }

  return `Ready: ${details.join(" · ")}. Backend validation still decides the final result.`;
}
