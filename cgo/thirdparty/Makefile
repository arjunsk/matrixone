GIT_COMMAND ?= git
FAISS_COMMIT ?= d87888b13e7eb339bb9c45825e9d20def6665171

all: faiss

.PHONY: faiss
faiss:
	if [ -e faiss ]; then \
		cd faiss && ${GIT_COMMAND} fetch origin && git reset --hard  $(FAISS_COMMIT); \
	else \
		${GIT_COMMAND} clone https://github.com/facebookresearch/faiss.git && cd faiss && git reset --hard  $(FAISS_COMMIT); \
	fi
	cd faiss && \
		export CMAKE_PREFIX_PATH=/opt/homebrew/opt/openblas:/opt/homebrew/opt/libomp:/opt/homebrew && \
		cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON -DFAISS_ENABLE_PYTHON=OFF . && \
		make -C build && \
		sudo make -C build install && \
		sudo cp build/c_api/libfaiss_c.dylib /usr/local/lib/libfaiss_c.dylib
