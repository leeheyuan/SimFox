import proj4 from "proj4";
import {Group} from "three"
import * as THREE from 'three'
import { Coordinates } from '@itowns/geographic'

function UTM48TOWGS84(x, y) {

    //const x = sumoX + (-410245.04);
    //const y = sumoY + (-3378526.74);

    proj4.defs("UTM48", "+proj=utm +zone=48 +ellps=WGS84 +datum=WGS84 +units=m +no_defs");
    return proj4("UTM48", "WGS84", [x, y]);
}

function TOTowns(lon, lat) {
    let coord = new itowns.Coordinates('EPSG:4326', lon, lat, 0);
    return coord
}


function createFlatPlaneAtCoord(lon, lat, size = 1000, altitude = 0, material) {
    const coord = new Coordinates('EPSG:4326', lon, lat, altitude)
    const position = coord.as('EPSG:4978')
    const posVec = new THREE.Vector3().copy(position)
    const up = posVec.clone().normalize()
    const coordEast = new Coordinates('EPSG:4326', lon + 0.0001, lat, altitude)
    const eastPos = coordEast.as('EPSG:4978');
    const east = new THREE.Vector3().subVectors(eastPos, position).normalize()
    const north = new THREE.Vector3().crossVectors(up, east).normalize()
    const correctedEast = new THREE.Vector3().crossVectors(north, up).normalize()
    const matrix = new THREE.Matrix4()
    matrix.makeBasis(correctedEast, north, up)
    const quaternion = new THREE.Quaternion().setFromRotationMatrix(matrix)

    const defaultMaterial = new THREE.MeshBasicMaterial({
        color: 0xff00ff,
        side: THREE.DoubleSide,
        transparent: false,
        opacity: 1,
        name: 'RoadLayerMaterial',
    })

  

    const geometry = new THREE.PlaneGeometry(size, size);
    const mesh = new THREE.Mesh(geometry, material || defaultMaterial)
    mesh.position.copy(position)
    mesh.quaternion.copy(quaternion) 
    return mesh
} 