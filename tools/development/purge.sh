#!/bin/bash
export KUBECONFIG=../../testbed/kind/consumer/config

kubectl delete solvers.nodecore.fluidos.eu -n fluidos --all
kubectl delete discoveries.advertisement.fluidos.eu -n fluidos --all
kubectl delete reservations.reservation.fluidos.eu -n fluidos --all
kubectl delete contracts.reservation.fluidos.eu -n fluidos --all
kubectl delete peeringcandidates.advertisement.fluidos.eu -n fluidos --all
kubectl delete transactions.reservation.fluidos.eu -n fluidos --all
kubectl delete allocations.nodecore.fluidos.eu -n fluidos --all
kubectl delete flavours.nodecore.fluidos.eu -n fluidos --all
kubectl delete pod -n fluidos --all

export KUBECONFIG=../../testbed/kind/provider/config

kubectl delete solvers.nodecore.fluidos.eu -n fluidos --all
kubectl delete discoveries.advertisement.fluidos.eu -n fluidos --all
kubectl delete reservations.reservation.fluidos.eu -n fluidos --all
kubectl delete contracts.reservation.fluidos.eu -n fluidos --all
kubectl delete peeringcandidates.advertisement.fluidos.eu -n fluidos --all
kubectl delete transactions.reservation.fluidos.eu -n fluidos --all
kubectl delete allocations.nodecore.fluidos.eu -n fluidos --all
kubectl delete flavours.nodecore.fluidos.eu -n fluidos --all
kubectl delete pod -n fluidos --all