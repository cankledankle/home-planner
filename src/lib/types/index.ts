export interface User {
	id: string;
	name: string;
	email: string;
	role: 'admin' | 'editor';
	created_at?: string;
	updated_at?: string;
}

export interface Plan {
	id: string;
	name: string;
	slug: string;
	type?: 'single_level' | 'multi_level';
	style?: 'cabin' | 'lodge' | 'modern' | 'ranch' | 'farmhouse';
	status: 'complete' | 'incomplete' | 'flagged';
	beds?: number;
	baths?: number;
	half_baths?: number;
	main_sf?: number;
	upper_sf?: number;
	lower_sf?: number;
	porch_deck_sf?: number;
	garage_sf?: number;
	garage_apartment_sf?: number;
	unfinished_sf?: number;
	heated_sf?: number;
	total_sf?: number;
	notes?: string;
	created_at?: string;
	updated_at?: string;
	created_by?: UserRef;
	updated_by?: UserRef;
}

export interface UserRef {
	id: string;
	name: string;
}

export interface File {
	id: string;
	plan_id: string;
	category: 'website' | 'reference' | 'technical' | '3d' | 'other';
	slot?: string;
	filename: string;
	storage_key: string;
	file_type: string;
	size_bytes: number;
	uploaded_at: string;
	uploaded_by?: UserRef;
}

export interface PlanFiles {
	website: {
		'render-front'?: File | null;
		'elevation-front'?: File | null;
		'elevation-left'?: File | null;
		'elevation-rear'?: File | null;
		'elevation-right'?: File | null;
		'floor-plan-main'?: File | null;
		'floor-plan-upper'?: File | null;
		'floor-plan-lower'?: File | null;
		poster?: File | null;
	};
	reference: File[];
	technical: File[];
	'3d': File[];
	other: File[];
}

export interface PlanWithFiles extends Plan {
	files: PlanFiles;
}

export interface Activity {
	id: string;
	user?: ActivityUser;
	plan?: ActivityPlan;
	action: string;
	detail?: Record<string, unknown>;
	created_at: string;
}

export interface ActivityUser {
	id: string;
	name: string;
}

export interface ActivityPlan {
	id: string;
	name: string;
}

export interface PaginatedResponse<T> {
	data: T[];
	meta: {
		page: number;
		limit: number;
		total: number;
		total_pages: number;
	};
}

export interface LoginCredentials {
	email: string;
	password: string;
}

export interface ApiError {
	error: {
		code: string;
		message: string;
	};
}

export interface SFTPCredentials {
	username: string;
	password?: string;
	host: string;
	port: number;
	permission: 'read' | 'readwrite';
}

export interface SFTPStatus {
	configured: boolean;
	healthy?: boolean;
	message: string;
}
