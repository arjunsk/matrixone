#include <stdlib.h>

// C function to calculate the squared L2 distance between two vectors of floats
float L2DistanceSqFloat32(int dim, float *ax, float *bx) {
    float distance = 0.0;
    for (int i = 0; i < dim; i++) {
        float diff = ax[i] - bx[i];
        distance += diff * diff;
    }
    return distance;
}
// C function to calculate the squared L2 distance between two vectors of doubles
double L2DistanceSqFloat64(int dim, double *ax, double *bx) {
    double distance = 0.0;
    for (int i = 0; i < dim; i++) {
        double diff = ax[i] - bx[i];
        distance += diff * diff;
    }
    return distance;
}